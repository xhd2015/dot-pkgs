package main

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/xhd2015/less-flags"
	"golang.design/x/clipboard"
)

const help = `
Usage: get-clipboard [OPTIONS]

Read system clipboard content:
  - Plain text is printed to stdout
  - Image, SVG, HTML, and other binary formats are saved to a file

Options:
  -o, --output FILE   output file path (by default a timestamped name with auto extension)
  -h, --help          show this help message
`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "get-clipboard: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	return runWithOutput(args, os.Stdout)
}

func runWithOutput(args []string, out io.Writer) error {
	var output string
	_, err := lessflags.String("-o,--output", &output).
		Help("-h,--help", help).
		Parse(args)
	if err != nil {
		return err
	}

	err = clipboard.Init()
	if err != nil {
		return fmt.Errorf("init clipboard: %w", err)
	}

	knownUTIs := []clipboard.Format{
		clipboard.Register("public.svg"),
		clipboard.Register("public.html"),
		clipboard.Register("public.rtf"),
		clipboard.Register("com.adobe.pdf"),
	}

	imgData := clipboard.Read(clipboard.FmtImage)
	if len(imgData) > 0 {
		ext := detectImageFormat(imgData)
		if ext == "" {
			return fmt.Errorf("cannot determine image format (magic bytes unrecognized)")
		}
		filename := makeOutputPath(output, imgData, ext)
		if err := os.WriteFile(filename, imgData, 0644); err != nil {
			return fmt.Errorf("write image: %w", err)
		}
		fmt.Fprintln(out, filename)
		return nil
	}

	textData := clipboard.Read(clipboard.FmtText)
	if len(textData) > 0 {
		fmt.Fprint(out, string(textData))
		return nil
	}

	fmts := clipboard.Formats()
	for _, f := range knownUTIs {
		fmts = append(fmts, f)
	}
	for _, f := range fmts {
		if f == clipboard.FmtText || f == clipboard.FmtImage {
			continue
		}
		data := clipboard.Read(f)
		if len(data) == 0 {
			continue
		}

		ext := extFromMIME(f.MIME())
		writeData := data
		if ext == "html" {
			if svgData, err := extractSVGFromHTML(data); err != nil {
				return err
			} else if svgData != nil {
				writeData = svgData
				ext = "svg"
			}
		}

		filename := makeOutputPath(output, writeData, ext)
		if err := os.WriteFile(filename, writeData, 0644); err != nil {
			return fmt.Errorf("write %s: %w", ext, err)
		}
		fmt.Fprintln(out, filename)
		return nil
	}

	names := make([]string, 0, len(fmts))
	for _, f := range fmts {
		mime := f.MIME()
		if mime != "" {
			names = append(names, mime)
		}
	}
	if len(names) > 0 {
		return fmt.Errorf("clipboard contains unsupported content (available: %s)", strings.Join(names, ", "))
	}
	return fmt.Errorf("clipboard contains unsupported content")
}

var svgDataURIPattern = regexp.MustCompile(`^<img\s+[^>]*\bsrc="data:image/svg\+xml;base64,([^"]*)"[^>]*>$`)
var bodyOpenPattern = regexp.MustCompile(`<body[^>]*>`)

func extractSVGFromHTML(htmlData []byte) ([]byte, error) {
	s := string(htmlData)

	startMatch := bodyOpenPattern.FindStringIndex(s)
	if startMatch == nil {
		return nil, nil
	}
	endIdx := strings.Index(s[startMatch[1]:], "</body>")
	if endIdx < 0 {
		return nil, nil
	}
	endIdx += startMatch[1]

	bodyContent := strings.TrimSpace(s[startMatch[1]:endIdx])

	matches := svgDataURIPattern.FindStringSubmatch(bodyContent)
	if matches == nil {
		return nil, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(matches[1])
	if err != nil {
		return nil, fmt.Errorf("decode base64 SVG: %w", err)
	}
	return decoded, nil
}

func extFromMIME(mime string) string {
	if i := strings.LastIndexByte(mime, '/'); i >= 0 {
		t := mime[i+1:]
		if j := strings.IndexByte(t, ';'); j >= 0 {
			t = t[:j]
		}
		if t != "" {
			return t
		}
	}
	if i := strings.LastIndexByte(mime, '.'); i >= 0 {
		t := mime[i+1:]
		if t != "" {
			return t
		}
	}
	return "bin"
}

func makeOutputPath(output string, data []byte, ext string) string {
	if output != "" {
		return output + "." + ext
	}
	return "/tmp/" + generateFilename(data, ext)
}

func detectImageFormat(data []byte) string {
	if len(data) < 4 {
		return ""
	}
	if data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' && data[3] == 'G' {
		return "png"
	}
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "jpg"
	}
	if len(data) >= 6 &&
		data[0] == 'G' && data[1] == 'I' && data[2] == 'F' && data[3] == '8' &&
		(data[4] == '7' || data[4] == '9') && data[5] == 'a' {
		return "gif"
	}
	if data[0] == 'B' && data[1] == 'M' {
		return "bmp"
	}
	if data[0] == 0x49 && data[1] == 0x49 && data[2] == 0x2A && data[3] == 0x00 {
		return "tiff"
	}
	if data[0] == 0x4D && data[1] == 0x4D && data[2] == 0x00 && data[3] == 0x2A {
		return "tiff"
	}
	return ""
}

func generateFilename(data []byte, ext string) string {
	ts := time.Now().Format("2006-01-02-15-04-05")
	h := md5.Sum(data)
	hash := fmt.Sprintf("%x", h)[:8]
	return fmt.Sprintf("%s-clipboard-%s.%s", ts, hash, ext)
}
