// Package getclipboard reads the system clipboard for peek (no write) and dump (write file).
package getclipboard

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Kind classifies clipboard content.
type Kind string

const (
	KindEmpty       Kind = "empty"
	KindText        Kind = "text"
	KindImage       Kind = "image"
	KindHTML        Kind = "html"
	KindSVG         Kind = "svg"
	KindRTF         Kind = "rtf"
	KindPDF         Kind = "pdf"
	KindOther       Kind = "other"
	KindUnsupported Kind = "unsupported"
)

// DefaultPreviewMax is the default peek preview character limit for text-like content.
const DefaultPreviewMax = 480

// maxNameAttempts caps collision suffixes when resolving --name paths.
const maxNameAttempts = 10000

// Format describes one clipboard format entry.
type Format struct {
	MIME string
	Read func() []byte
}

// Source abstracts clipboard reads (injectable in tests).
type Source interface {
	Init() error
	ReadImage() []byte
	ReadText() []byte
	// ExtraFormats returns non-text/image formats to try (SVG/HTML/RTF/PDF/…).
	ExtraFormats() []Format
}

// Content is clipboard payload ready for peek or dump.
type Content struct {
	Kind      Kind
	Ext       string
	Data      []byte
	MIME      string
	Available []string // populated when KindUnsupported
}

// PeekResult is a non-destructive summary of clipboard content.
type PeekResult struct {
	Kind      Kind     `json:"kind"`
	Ext       string   `json:"ext,omitempty"`
	Bytes     int      `json:"bytes"`
	Preview   string   `json:"preview,omitempty"`
	MIME      string   `json:"mime,omitempty"`
	Available []string `json:"available,omitempty"`
}

// DumpOptions controls where dump writes files.
type DumpOptions struct {
	// Output is a path stem; extension is appended (like CLI -o).
	Output string
	// Name is a basename under Dir (like CLI -n).
	Name string
	// Dir defaults to /tmp for timestamped and --name paths.
	Dir string
}

// DumpResult is the path written by Dump.
type DumpResult struct {
	Path  string `json:"path"`
	Kind  Kind   `json:"kind"`
	Ext   string `json:"ext"`
	Bytes int    `json:"bytes"`
}

// Read loads clipboard content without writing files.
func Read(src Source) (*Content, error) {
	if src == nil {
		src = System()
	}
	if err := src.Init(); err != nil {
		return nil, fmt.Errorf("init clipboard: %w", err)
	}

	if img := src.ReadImage(); len(img) > 0 {
		ext := DetectImageFormat(img)
		if ext == "" {
			return nil, fmt.Errorf("cannot determine image format (magic bytes unrecognized)")
		}
		return &Content{Kind: KindImage, Ext: ext, Data: img, MIME: "image/" + ext}, nil
	}

	if text := src.ReadText(); len(text) > 0 {
		return &Content{Kind: KindText, Ext: "txt", Data: text, MIME: "text/plain"}, nil
	}

	fmts := src.ExtraFormats()
	var available []string
	for _, f := range fmts {
		if f.MIME != "" {
			available = append(available, f.MIME)
		}
		data := f.Read()
		if len(data) == 0 {
			continue
		}
		ext := ExtFromMIME(f.MIME)
		writeData := data
		kind := kindFromExt(ext)
		if ext == "html" {
			if svgData, err := ExtractSVGFromHTML(data); err != nil {
				return nil, err
			} else if svgData != nil {
				writeData = svgData
				ext = "svg"
				kind = KindSVG
			}
		}
		return &Content{Kind: kind, Ext: ext, Data: writeData, MIME: f.MIME}, nil
	}

	if len(available) > 0 {
		return &Content{Kind: KindUnsupported, Available: available}, nil
	}
	return &Content{Kind: KindEmpty}, nil
}

// Peek summarizes content without writing. previewMax <= 0 uses DefaultPreviewMax.
func Peek(src Source, previewMax int) (*PeekResult, error) {
	c, err := Read(src)
	if err != nil {
		return nil, err
	}
	return c.Peek(previewMax), nil
}

// Peek builds a PeekResult from Content.
func (c *Content) Peek(previewMax int) *PeekResult {
	if c == nil {
		return &PeekResult{Kind: KindEmpty}
	}
	if previewMax <= 0 {
		previewMax = DefaultPreviewMax
	}
	out := &PeekResult{
		Kind:      c.Kind,
		Ext:       c.Ext,
		Bytes:     len(c.Data),
		MIME:      c.MIME,
		Available: c.Available,
	}
	switch c.Kind {
	case KindText, KindHTML, KindRTF, KindSVG:
		out.Preview = TruncatePreview(string(c.Data), previewMax)
	case KindImage:
		out.Preview = fmt.Sprintf("%s image", c.Ext)
	case KindPDF:
		out.Preview = "PDF document"
	case KindOther:
		out.Preview = c.Ext
	case KindUnsupported:
		out.Preview = "unsupported"
	case KindEmpty:
		out.Preview = ""
	}
	return out
}

// Dump writes content to a file (including text as .txt) and returns the path.
func Dump(src Source, opts DumpOptions) (*DumpResult, error) {
	c, err := Read(src)
	if err != nil {
		return nil, err
	}
	return c.Dump(opts)
}

// Dump writes Content to a file.
func (c *Content) Dump(opts DumpOptions) (*DumpResult, error) {
	if c == nil || c.Kind == KindEmpty {
		return nil, fmt.Errorf("clipboard is empty")
	}
	if c.Kind == KindUnsupported {
		if len(c.Available) > 0 {
			return nil, fmt.Errorf("clipboard contains unsupported content (available: %s)", strings.Join(c.Available, ", "))
		}
		return nil, fmt.Errorf("clipboard contains unsupported content")
	}
	if opts.Output != "" && opts.Name != "" {
		return nil, fmt.Errorf("--name and --output are mutually exclusive")
	}
	name := opts.Name
	if name != "" {
		var err error
		name, err = ValidateName(name)
		if err != nil {
			return nil, err
		}
	}
	dir := opts.Dir
	if dir == "" {
		dir = "/tmp"
	}
	filename, err := MakeOutputPath(opts.Output, name, dir, c.Data, c.Ext)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filename, c.Data, 0644); err != nil {
		return nil, fmt.Errorf("write %s: %w", c.Ext, err)
	}
	return &DumpResult{Path: filename, Kind: c.Kind, Ext: c.Ext, Bytes: len(c.Data)}, nil
}

func kindFromExt(ext string) Kind {
	switch strings.ToLower(ext) {
	case "html":
		return KindHTML
	case "svg", "svg+xml":
		return KindSVG
	case "rtf":
		return KindRTF
	case "pdf":
		return KindPDF
	case "txt", "plain":
		return KindText
	default:
		return KindOther
	}
}

// TruncatePreview collapses whitespace runs lightly and truncates to max runes.
func TruncatePreview(text string, max int) string {
	if max <= 0 {
		return ""
	}
	s := strings.TrimSpace(text)
	if s == "" {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "…"
}

var svgDataURIPattern = regexp.MustCompile(`^<img\s+[^>]*\bsrc="data:image/svg\+xml;base64,([^"]*)"[^>]*>$`)
var bodyOpenPattern = regexp.MustCompile(`<body[^>]*>`)

// ExtractSVGFromHTML pulls a lone base64 SVG data-URI img from an HTML body, if present.
func ExtractSVGFromHTML(htmlData []byte) ([]byte, error) {
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

// ExtFromMIME maps a MIME or UTI string to a file extension.
func ExtFromMIME(mime string) string {
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

// ValidateName ensures --name is a safe basename.
func ValidateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("--name must be a non-empty basename (no path separators)")
	}
	if name == "." || name == ".." ||
		strings.Contains(name, "/") || strings.Contains(name, "\\") ||
		strings.Contains(name, string(filepath.Separator)) {
		return "", fmt.Errorf("--name must be a non-empty basename (no path separators)")
	}
	return name, nil
}

// MakeOutputPath builds the dump path. Dir is used for timestamped and named paths.
func MakeOutputPath(output, name, dir string, data []byte, ext string) (string, error) {
	if output != "" {
		return output + "." + ext, nil
	}
	if dir == "" {
		dir = "/tmp"
	}
	if name != "" {
		return UniqueNamedPath(dir, name, ext)
	}
	return filepath.Join(dir, GenerateFilename(data, ext)), nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// UniqueNamedPath returns dir/stem.ext, or dir/stem-1.ext, stem-2.ext, ... if taken.
func UniqueNamedPath(dir, stem, ext string) (string, error) {
	candidate := filepath.Join(dir, stem+"."+ext)
	if !fileExists(candidate) {
		return candidate, nil
	}
	for i := 1; i <= maxNameAttempts; i++ {
		candidate = filepath.Join(dir, fmt.Sprintf("%s-%d.%s", stem, i, ext))
		if !fileExists(candidate) {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not find free path for --name %q under %s (tried up to -%d)", stem, dir, maxNameAttempts)
}

// DetectImageFormat returns png/jpg/gif/bmp/tiff from magic bytes, or "".
func DetectImageFormat(data []byte) string {
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

// GenerateFilename builds <ts>-clipboard-<hash8>.<ext>.
func GenerateFilename(data []byte, ext string) string {
	ts := time.Now().Format("2006-01-02-15-04-05")
	h := md5.Sum(data)
	hash := fmt.Sprintf("%x", h)[:8]
	return fmt.Sprintf("%s-clipboard-%s.%s", ts, hash, ext)
}
