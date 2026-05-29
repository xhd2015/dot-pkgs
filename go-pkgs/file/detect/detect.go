package detect

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func DetectFileType(path string) (string, bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", false, err
	}
	if stat.IsDir() {
		return "", false, nil
	}
	if stat.Size() == 0 {
		return "", false, nil
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", false, err
	}
	buf = buf[:n]

	if desc, ok := detectByMagic(buf); ok {
		if isBinaryMagicDesc(desc) {
			return desc, true, nil
		}
		return desc, false, nil
	}

	if isBinary, desc := detectByContentSniff(buf); desc != "" {
		return desc, isBinary, nil
	}

	return "", false, nil
}

func isBinaryMagicDesc(desc string) bool {
	return desc != "text"
}

func detectByMagic(buf []byte) (string, bool) {
	if len(buf) < 4 {
		return "", false
	}

	be := binary.BigEndian
	le := binary.LittleEndian

	magic4be := be.Uint32(buf[:4])
	magic4le := le.Uint32(buf[:4])

	if s := detectMachO(buf, magic4be, magic4le, be, le); s != "" {
		return s, true
	}

	if s := detectELF(buf, magic4be); s != "" {
		return s, true
	}

	if s := detectPE(buf, magic4le); s != "" {
		return s, true
	}

	if s := detectArchive(buf, magic4be, magic4le); s != "" {
		return s, true
	}

	if len(buf) >= 2 {
		if s := detectImageAudioVideo(buf, be, le); s != "" {
			return s, true
		}
	}

	if s, ok := detectTextLike(buf); ok {
		return s, true
	}

	return "", false
}

func detectByContentSniff(buf []byte) (bool, string) {
	hasNull := false
	for _, b := range buf {
		if b == 0 {
			hasNull = true
			break
		}
	}
	if hasNull {
		return true, "binary file"
	}

	if buf[0] >= 0x80 {
		return true, "binary file"
	}

	return false, "text file"
}

// -- Mach-O --

func detectMachO(buf []byte, magic4be, magic4le uint32, be, le binary.ByteOrder) string {
	const (
		MH_MAGIC    = 0xFEEDFACE
		MH_CIGAM    = 0xCEFAEDFE
		MH_MAGIC_64 = 0xFEEDFACF
		MH_CIGAM_64 = 0xCFFAEDFE
		FAT_MAGIC   = 0xCAFEBABE
		FAT_CIGAM   = 0xBEBAFECA
	)

	var is64 bool
	var bo binary.ByteOrder
	switch magic4be {
	case MH_MAGIC:
		bo = be
	case MH_MAGIC_64:
		bo = be
		is64 = true
	case MH_CIGAM:
		bo = le
	case MH_CIGAM_64:
		bo = le
		is64 = true
	case FAT_MAGIC:
		return detectFatMachO(buf, be)
	case FAT_CIGAM:
		return detectFatMachO(buf, le)
	default:
		return ""
	}

	return describeMachO(buf, bo, is64)
}

func describeMachO(buf []byte, order binary.ByteOrder, is64 bool) string {
	if len(buf) < 8 {
		if is64 {
			return "Mach-O 64-bit executable"
		}
		return "Mach-O executable"
	}

	cpuType := order.Uint32(buf[4:8])
	cpu := machOCPUName(cpuType)

	bits := ""
	if is64 {
		bits = " 64-bit"
	}
	return fmt.Sprintf("Mach-O%s executable %s", bits, cpu)
}

func detectFatMachO(buf []byte, order binary.ByteOrder) string {
	if len(buf) < 8 {
		return "Mach-O universal binary"
	}
	nfat := order.Uint32(buf[4:8])
	return fmt.Sprintf("Mach-O universal binary (%d architectures)", nfat)
}

func machOCPUName(cpuType uint32) string {
	cpuNames := map[uint32]string{
		7:           "x86",
		0x01000007:  "x86_64",
		12:          "arm",
		0x0100000C:  "arm64",
		18:          "ppc",
		0x01000012:  "ppc64",
	}
	if name, ok := cpuNames[cpuType]; ok {
		return name
	}
	return fmt.Sprintf("cputype(%d)", cpuType)
}

// -- ELF --

func detectELF(buf []byte, magic4be uint32) string {
	if magic4be != 0x7F454C46 {
		return ""
	}
	if len(buf) < 20 {
		return "ELF executable"
	}

	class := buf[4]
	var byteOrder binary.ByteOrder = binary.BigEndian
	if buf[5] == 1 {
		byteOrder = binary.LittleEndian
	}
	machine := byteOrder.Uint16(buf[18:20])

	bits := "32-bit"
	if class == 2 {
		bits = "64-bit"
	}
	endian := "MSB"
	if buf[5] == 1 {
		endian = "LSB"
	}

	cpu := elfMachineName(machine)
	return fmt.Sprintf("ELF %s %s executable %s", bits, endian, cpu)
}

func elfMachineName(machine uint16) string {
	names := map[uint16]string{
		3:   "x86",
		62:  "x86_64",
		40:  "ARM",
		183: "AArch64",
		20:  "PowerPC",
		21:  "PowerPC64",
		8:   "MIPS",
	}
	if name, ok := names[machine]; ok {
		return name
	}
	return fmt.Sprintf("machine(%d)", machine)
}

// -- Windows PE --

func detectPE(buf []byte, magic4le uint32) string {
	if magic4le != 0x00004550 {
		return ""
	}
	if len(buf) < 6 {
		return "PE executable"
	}

	machine := binary.LittleEndian.Uint16(buf[4:6])
	cpu := peMachineName(machine)

	var pe32Plus bool
	if len(buf) >= 0x18+2 {
		magic := binary.LittleEndian.Uint16(buf[0x18:])
		if magic == 0x020B {
			pe32Plus = true
		}
	}

	bits := "32-bit"
	if pe32Plus {
		bits = "64-bit"
	}

	return fmt.Sprintf("PE %s executable %s", bits, cpu)
}

func peMachineName(machine uint16) string {
	names := map[uint16]string{
		0x014C: "x86",
		0x8664: "x86_64",
		0x01C0: "ARM",
		0xAA64: "AArch64",
		0x0200: "IA64",
	}
	if name, ok := names[machine]; ok {
		return name
	}
	return fmt.Sprintf("machine(0x%X)", machine)
}

// -- Archives & documents --

func detectArchive(buf []byte, magic4be, magic4le uint32) string {
	switch {
	case magic4be&0xFFFFFF00 == 0x1F8B0800:
		return "gzip compressed data"
	case magic4le == 0x04034B50:
		return "Zip archive"
	case matchBytes(buf, 0, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}):
		return "XZ compressed data"
	case matchBytes(buf, 0, []byte("BZh")):
		return "bzip2 compressed data"
	case matchBytes(buf, 0, []byte{0x52, 0x61, 0x72, 0x21, 0x1A, 0x07, 0x00}):
		return "RAR archive"
	case matchBytes(buf, 0, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27, 0x1C}):
		return "7-zip archive"
	case matchBytes(buf, 0, []byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x00, 0x30, 0x30}):
		return "tar archive"
	case matchBytes(buf, 0, []byte{0x75, 0x73, 0x74, 0x61, 0x72, 0x20, 0x20, 0x00}):
		return "tar archive"
	case matchBytes(buf, 0, []byte{0x5A, 0x4D, 0x41, 0x50}):
		return "Z compressed data"
	case magic4le == 0x25215053:
		return "PostScript document"
	case magic4le == 0x46445025:
		return "PDF document"
	case matchBytes(buf, 0, []byte{0xCA, 0xFE, 0xBA, 0xBE}):
		return "Java class file"
	case matchBytes(buf, 0, []byte("wasm")) || matchBytes(buf, 0, []byte{0x00, 0x61, 0x73, 0x6D}):
		return "WebAssembly binary"
	case matchBytes(buf, 0, []byte("SQLite format 3\x00")):
		return "SQLite database"
	}
	return ""
}

// -- Images, audio, video, fonts --

func detectImageAudioVideo(buf []byte, be, le binary.ByteOrder) string {
	magic2be := be.Uint16(buf[:2])

	if len(buf) >= 8 {
		magic8be := be.Uint64(buf[:8])
		switch {
		case matchBytes(buf, 0, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
			return "PNG image"
		case magic8be&0xFFFFFFFFFFFF0000 == 0x4749463839610000:
			return "GIF image"
		case be.Uint32(buf[:4]) == 0x52494646 && matchBytes(buf, 8, []byte("WEBP")):
			return "WebP image"
		case be.Uint32(buf[:4]) == 0x00000100 && le.Uint16(buf[2:4]) >= 1:
			return "ICO icon"
		}
	}

	switch {
	case magic2be == 0xFFD8:
		return "JPEG image"
	case magic2be == 0x424D:
		return "BMP image"
	case magic2be == 0x4D4D || magic2be == 0x4949:
		return "TIFF image"
	case matchBytes(buf, 0, []byte{0x00, 0x00, 0x01, 0x00}):
		return "Windows icon"
	case matchBytes(buf, 0, []byte{0xFF, 0xFB}) || matchBytes(buf, 0, []byte{0xFF, 0xF3}) || matchBytes(buf, 0, []byte{0xFF, 0xF2}):
		return "MP3 audio"
	case matchBytes(buf, 0, []byte("ID3")):
		return "MP3 audio"
	case matchBytes(buf, 0, []byte("RIFF")) && len(buf) >= 12 && matchBytes(buf, 8, []byte("WAVE")):
		return "WAV audio"
	case matchBytes(buf, 0, []byte("fLaC")):
		return "FLAC audio"
	case matchBytes(buf, 0, []byte("OggS")):
		return "OGG container"
	case matchBytes(buf, 4, []byte("ftyp")):
		return "MP4 container"
	case matchBytes(buf, 0, []byte{0x1A, 0x45, 0xDF, 0xA3}):
		return "WebM container"
	case matchBytes(buf, 0, []byte("RIFF")) && len(buf) >= 12 && matchBytes(buf, 8, []byte("AVI ")):
		return "AVI video"
	case matchBytes(buf, 0, []byte{0x00, 0x01, 0x00, 0x00}):
		return "TrueType font"
	case matchBytes(buf, 0, []byte("OTTO")):
		return "OpenType font"
	case matchBytes(buf, 0, []byte{0x77, 0x4F, 0x46, 0x46}):
		return "WOFF font"
	case matchBytes(buf, 0, []byte{0x77, 0x4F, 0x46, 0x32}):
		return "WOFF2 font"
	}

	return ""
}

// -- Text detection --

func detectTextLike(buf []byte) (string, bool) {
	if len(buf) < 2 {
		return "", false
	}

	hasNull := false
	hasHighBit := false
	hasLowCtrl := false
	for _, b := range buf {
		if b == 0 {
			hasNull = true
		}
		if b > 0x7F {
			hasHighBit = true
		}
		if b < 0x20 && b != 0x09 && b != 0x0A && b != 0x0D {
			hasLowCtrl = true
		}
	}

	if hasNull {
		return "", false
	}

	if hasLowCtrl && hasHighBit {
		return "", false
	}

	if !hasHighBit && !hasLowCtrl {
		return "text", true
	}

	if hasHighBit && !hasLowCtrl {
		sample := make([]byte, 0, len(buf))
		for _, b := range buf {
			if b >= 0x20 || b == 0x09 || b == 0x0A || b == 0x0D {
				sample = append(sample, b)
			}
		}
		if looksLikeUTF8(sample) {
			return "text", true
		}
		return "", false
	}

	return "", false
}

// -- Helpers --

func matchBytes(buf []byte, offset int, prefix []byte) bool {
	if offset+len(prefix) > len(buf) {
		return false
	}
	for i, b := range prefix {
		if buf[offset+i] != b {
			return false
		}
	}
	return true
}

func looksLikeUTF8(data []byte) bool {
	for i := 0; i < len(data); {
		if data[i] < 0x80 {
			i++
			continue
		}
		seqLen := 0
		switch {
		case data[i]&0xE0 == 0xC0:
			seqLen = 2
		case data[i]&0xF0 == 0xE0:
			seqLen = 3
		case data[i]&0xF8 == 0xF0:
			seqLen = 4
		default:
			return false
		}
		if i+seqLen > len(data) {
			return false
		}
		for j := 1; j < seqLen; j++ {
			if data[i+j]&0xC0 != 0x80 {
				return false
			}
		}
		i += seqLen
	}
	return true
}
