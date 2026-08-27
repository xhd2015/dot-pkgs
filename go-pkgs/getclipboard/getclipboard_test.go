package getclipboard

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

type fakeSource struct {
	image   []byte
	text    []byte
	formats []Format
	initErr error
}

func (f *fakeSource) Init() error              { return f.initErr }
func (f *fakeSource) ReadImage() []byte        { return f.image }
func (f *fakeSource) ReadText() []byte         { return f.text }
func (f *fakeSource) ExtraFormats() []Format   { return f.formats }

func TestDetectImageFormat(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		expect string
	}{
		{name: "png", data: []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, expect: "png"},
		{name: "jpeg", data: []byte{0xFF, 0xD8, 0xFF, 0xE0}, expect: "jpg"},
		{name: "gif87a", data: []byte{'G', 'I', 'F', '8', '7', 'a'}, expect: "gif"},
		{name: "gif89a", data: []byte{'G', 'I', 'F', '8', '9', 'a'}, expect: "gif"},
		{name: "bmp", data: []byte{'B', 'M', 0x00, 0x00}, expect: "bmp"},
		{name: "tiff_little_endian", data: []byte{0x49, 0x49, 0x2A, 0x00}, expect: "tiff"},
		{name: "tiff_big_endian", data: []byte{0x4D, 0x4D, 0x00, 0x2A}, expect: "tiff"},
		{name: "too_short", data: []byte{0x00}, expect: ""},
		{name: "empty", data: []byte{}, expect: ""},
		{name: "unknown_bytes", data: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, expect: ""},
		{name: "partial_gif_too_short", data: []byte{'G', 'I', 'F', '8'}, expect: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectImageFormat(tt.data)
			if got != tt.expect {
				t.Errorf("DetectImageFormat() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	data := []byte("test image data")
	name := GenerateFilename(data, "png")
	pattern := `^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-clipboard-[0-9a-f]{8}\.png$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		t.Errorf("GenerateFilename() = %q, does not match %s", name, pattern)
	}
}

func TestGenerateFilenameConsistent(t *testing.T) {
	data := []byte("consistent data")
	if GenerateFilename(data, "jpg") != GenerateFilename(data, "jpg") {
		t.Error("same data should produce same hash")
	}
}

func TestGenerateFilenameDifferentData(t *testing.T) {
	if GenerateFilename([]byte("data A"), "png") == GenerateFilename([]byte("data B"), "png") {
		t.Error("different data should produce different names")
	}
}

func TestExtFromMIME(t *testing.T) {
	tests := []struct {
		name   string
		mime   string
		expect string
	}{
		{name: "html", mime: "text/html", expect: "html"},
		{name: "html_with_charset", mime: "text/html;charset=utf-8", expect: "html"},
		{name: "plain", mime: "text/plain", expect: "plain"},
		{name: "svg", mime: "image/svg+xml", expect: "svg+xml"},
		{name: "png", mime: "image/png", expect: "png"},
		{name: "pdf", mime: "application/pdf", expect: "pdf"},
		{name: "rtf", mime: "text/rtf", expect: "rtf"},
		{name: "jpeg", mime: "image/jpeg", expect: "jpeg"},
		{name: "json", mime: "application/json", expect: "json"},
		{name: "public_uti_svg", mime: "public.svg", expect: "svg"},
		{name: "public_uti_html", mime: "public.html", expect: "html"},
		{name: "public_uti_rtf", mime: "public.rtf", expect: "rtf"},
		{name: "adobe_pdf", mime: "com.adobe.pdf", expect: "pdf"},
		{name: "empty", mime: "", expect: "bin"},
		{name: "plain_with_charset", mime: "text/plain;charset=utf-8", expect: "plain"},
		{name: "javascript", mime: "text/javascript", expect: "javascript"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtFromMIME(tt.mime)
			if got != tt.expect {
				t.Errorf("ExtFromMIME(%q) = %q, want %q", tt.mime, got, tt.expect)
			}
		})
	}
}

func TestExtractSVGFromHTML(t *testing.T) {
	svgContent := `<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><rect fill="red" width="100" height="100"/></svg>`
	svgBase64 := base64.StdEncoding.EncodeToString([]byte(svgContent))
	wrapHTML := func(bodyInner string) string {
		return `<meta charset='utf-8'><html><head></head><body>` + bodyInner + `</body></html>`
	}
	tests := []struct {
		name      string
		html      string
		expectSVG string
		expectErr bool
	}{
		{name: "body_with_only_svg_img", html: wrapHTML(`<img src="data:image/svg+xml;base64,` + svgBase64 + `">`), expectSVG: svgContent},
		{name: "body_with_img_other_attrs", html: wrapHTML(`<img class="diagram" src="data:image/svg+xml;base64,` + svgBase64 + `" alt="diagram">`), expectSVG: svgContent},
		{name: "body_with_img_and_whitespace", html: wrapHTML("\n\t\t\t<img src=\"data:image/svg+xml;base64," + svgBase64 + "\">\n\t\t"), expectSVG: svgContent},
		{name: "body_with_img_and_other_content", html: wrapHTML(`<img src="data:image/svg+xml;base64,` + svgBase64 + `"><p>text</p>`)},
		{name: "img_with_png_data_uri", html: wrapHTML(`<img src="data:image/png;base64,` + svgBase64 + `">`)},
		{name: "body_without_img", html: wrapHTML(`<p>hello</p>`)},
		{name: "no_body_tag", html: `<img src="data:image/svg+xml;base64,` + svgBase64 + `">`},
		{name: "invalid_base64", html: wrapHTML(`<img src="data:image/svg+xml;base64,!!!">`), expectErr: true},
		{name: "empty_html", html: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractSVGFromHTML([]byte(tt.html))
			if tt.expectErr {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.expectSVG == "" {
				if got != nil {
					t.Errorf("expected nil, got %q", string(got))
				}
				return
			}
			if string(got) != tt.expectSVG {
				t.Errorf("SVG mismatch:\ngot:  %q\nwant: %q", string(got), tt.expectSVG)
			}
		})
	}
}

func TestMakeOutputPath(t *testing.T) {
	data := []byte("test")
	ext := "png"
	t.Run("with_output", func(t *testing.T) {
		got, err := MakeOutputPath("out", "", "/tmp", data, ext)
		if err != nil {
			t.Fatal(err)
		}
		if got != "out.png" {
			t.Errorf("got %q, want out.png", got)
		}
	})
	t.Run("without_output", func(t *testing.T) {
		dir := t.TempDir()
		got, err := MakeOutputPath("", "", dir, data, ext)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(got, dir+string(filepath.Separator)) {
			t.Errorf("got %q, want under %s", got, dir)
		}
		pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-clipboard-[0-9a-f]{8}\.png$`)
		if !pattern.MatchString(filepath.Base(got)) {
			t.Errorf("base %q does not match pattern", filepath.Base(got))
		}
	})
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{name: "ok", in: "some-meaningful", want: "some-meaningful"},
		{name: "trim", in: "  demo  ", want: "demo"},
		{name: "empty", in: "", wantErr: true},
		{name: "spaces", in: "   ", wantErr: true},
		{name: "slash", in: "a/b", wantErr: true},
		{name: "backslash", in: `a\b`, wantErr: true},
		{name: "dot", in: ".", wantErr: true},
		{name: "dotdot", in: "..", wantErr: true},
		{name: "with_ext_as_stem", in: "foo.png", want: "foo.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateName(tt.in)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %q", got)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("ValidateName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestUniqueNamedPath(t *testing.T) {
	dir := t.TempDir()
	stem := "demo"
	ext := "png"
	t.Run("free", func(t *testing.T) {
		got, err := UniqueNamedPath(dir, stem, ext)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "demo.png")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
	t.Run("base_taken", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(dir, "demo.png"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := UniqueNamedPath(dir, stem, ext)
		if err != nil {
			t.Fatal(err)
		}
		if got != filepath.Join(dir, "demo-1.png") {
			t.Errorf("got %q", got)
		}
	})
}

func TestPeekAndDumpText(t *testing.T) {
	src := &fakeSource{text: []byte("hello clipboard")}
	peek, err := Peek(src, 10)
	if err != nil {
		t.Fatal(err)
	}
	if peek.Kind != KindText || peek.Bytes != 15 || peek.Preview != "hello clip…" {
		t.Fatalf("peek = %+v", peek)
	}
	dir := t.TempDir()
	dump, err := Dump(src, DumpOptions{Dir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if dump.Kind != KindText || dump.Ext != "txt" {
		t.Fatalf("dump = %+v", dump)
	}
	data, err := os.ReadFile(dump.Path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello clipboard" {
		t.Fatalf("file = %q", data)
	}
}

func TestPeekImageAndEmpty(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	peek, err := Peek(&fakeSource{image: png}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if peek.Kind != KindImage || peek.Ext != "png" || peek.Preview != "png image" {
		t.Fatalf("peek = %+v", peek)
	}
	empty, err := Peek(&fakeSource{}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if empty.Kind != KindEmpty {
		t.Fatalf("empty kind = %q", empty.Kind)
	}
}

func TestDumpEmptyErrors(t *testing.T) {
	_, err := Dump(&fakeSource{}, DumpOptions{Dir: t.TempDir()})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("err = %v", err)
	}
}
