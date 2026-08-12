package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestDetectImageFormat(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		expect string
	}{
		{
			name:   "png",
			data:   []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A},
			expect: "png",
		},
		{
			name:   "jpeg",
			data:   []byte{0xFF, 0xD8, 0xFF, 0xE0},
			expect: "jpg",
		},
		{
			name:   "gif87a",
			data:   []byte{'G', 'I', 'F', '8', '7', 'a'},
			expect: "gif",
		},
		{
			name:   "gif89a",
			data:   []byte{'G', 'I', 'F', '8', '9', 'a'},
			expect: "gif",
		},
		{
			name:   "bmp",
			data:   []byte{'B', 'M', 0x00, 0x00},
			expect: "bmp",
		},
		{
			name:   "tiff_little_endian",
			data:   []byte{0x49, 0x49, 0x2A, 0x00},
			expect: "tiff",
		},
		{
			name:   "tiff_big_endian",
			data:   []byte{0x4D, 0x4D, 0x00, 0x2A},
			expect: "tiff",
		},
		{
			name:   "too_short",
			data:   []byte{0x00},
			expect: "",
		},
		{
			name:   "empty",
			data:   []byte{},
			expect: "",
		},
		{
			name:   "unknown_bytes",
			data:   []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expect: "",
		},
		{
			name:   "partial_gif_too_short",
			data:   []byte{'G', 'I', 'F', '8'},
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectImageFormat(tt.data)
			if got != tt.expect {
				t.Errorf("detectImageFormat() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestGenerateFilename(t *testing.T) {
	data := []byte("test image data")
	name := generateFilename(data, "png")

	pattern := `^\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-clipboard-[0-9a-f]{8}\.png$`
	matched, err := regexp.MatchString(pattern, name)
	if err != nil {
		t.Fatal(err)
	}
	if !matched {
		t.Errorf("generateFilename() = %q, does not match pattern %s", name, pattern)
	}
}

func TestGenerateFilenameConsistent(t *testing.T) {
	data := []byte("consistent data")
	name1 := generateFilename(data, "jpg")
	name2 := generateFilename(data, "jpg")
	if name1 != name2 {
		t.Errorf("same data should produce same hash: %q vs %q", name1, name2)
	}
}

func TestGenerateFilenameDifferentData(t *testing.T) {
	name1 := generateFilename([]byte("data A"), "png")
	name2 := generateFilename([]byte("data B"), "png")
	if name1 == name2 {
		t.Errorf("different data should produce different names")
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
			got := extFromMIME(tt.mime)
			if got != tt.expect {
				t.Errorf("extFromMIME(%q) = %q, want %q", tt.mime, got, tt.expect)
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
		{
			name:      "body_with_only_svg_img",
			html:      wrapHTML(`<img src="data:image/svg+xml;base64,` + svgBase64 + `">`),
			expectSVG: svgContent,
		},
		{
			name:      "body_with_img_other_attrs",
			html:      wrapHTML(`<img class="diagram" src="data:image/svg+xml;base64,` + svgBase64 + `" alt="diagram">`),
			expectSVG: svgContent,
		},
		{
			name:      "body_with_img_and_whitespace",
			html:      wrapHTML(`
			<img src="data:image/svg+xml;base64,` + svgBase64 + `">
		`),
			expectSVG: svgContent,
		},
		{
			name: "body_with_img_and_other_content",
			html: wrapHTML(`<img src="data:image/svg+xml;base64,` + svgBase64 + `"><p>text</p>`),
		},
		{
			name: "img_with_png_data_uri",
			html: wrapHTML(`<img src="data:image/png;base64,` + svgBase64 + `">`),
		},
		{
			name: "body_without_img",
			html: wrapHTML(`<p>hello</p>`),
		},
		{
			name: "no_body_tag",
			html: `<img src="data:image/svg+xml;base64,` + svgBase64 + `">`,
		},
		{
			name:      "invalid_base64",
			html:      wrapHTML(`<img src="data:image/svg+xml;base64,!!!">`),
			expectErr: true,
		},
		{
			name: "empty_html",
			html: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractSVGFromHTML([]byte(tt.html))
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
			if got == nil {
				t.Errorf("expected SVG data, got nil")
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
		got, err := makeOutputPath("out", "", data, ext)
		if err != nil {
			t.Fatal(err)
		}
		if got != "out.png" {
			t.Errorf("makeOutputPath = %q, want %q", got, "out.png")
		}
	})

	t.Run("with_output_path", func(t *testing.T) {
		got, err := makeOutputPath("/some/dir/file", "", data, ext)
		if err != nil {
			t.Fatal(err)
		}
		if got != "/some/dir/file.png" {
			t.Errorf("makeOutputPath = %q, want %q", got, "/some/dir/file.png")
		}
	})

	t.Run("without_output", func(t *testing.T) {
		got, err := makeOutputPath("", "", data, ext)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(got, "/tmp/") {
			t.Errorf("makeOutputPath = %q, should start with /tmp/", got)
		}
		pattern := `^/tmp/\d{4}-\d{2}-\d{2}-\d{2}-\d{2}-\d{2}-clipboard-[0-9a-f]{8}\.png$`
		if !regexp.MustCompile(pattern).MatchString(got) {
			t.Errorf("makeOutputPath = %q, does not match pattern %s", got, pattern)
		}
	})

	t.Run("with_name_free", func(t *testing.T) {
		dir := t.TempDir()
		// uniqueNamedPath is used under /tmp; exercise via uniqueNamedPath directly
		// and keep makeOutputPath name branch covered when /tmp is free enough.
		got, err := uniqueNamedPath(dir, "some-meaningful", "png")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "some-meaningful.png")
		if got != want {
			t.Errorf("uniqueNamedPath = %q, want %q", got, want)
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
			got, err := validateName(tt.in)
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
				t.Errorf("validateName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestUniqueNamedPath(t *testing.T) {
	dir := t.TempDir()
	stem := "demo"
	ext := "png"

	t.Run("free", func(t *testing.T) {
		got, err := uniqueNamedPath(dir, stem, ext)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "demo.png")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("base_taken", func(t *testing.T) {
		base := filepath.Join(dir, "demo.png")
		if err := os.WriteFile(base, []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := uniqueNamedPath(dir, stem, ext)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "demo-1.png")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("base_and_1_taken", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(dir, "demo-1.png"), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
		got, err := uniqueNamedPath(dir, stem, ext)
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "demo-2.png")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("jpg_ext", func(t *testing.T) {
		got, err := uniqueNamedPath(dir, "shot", "jpg")
		if err != nil {
			t.Fatal(err)
		}
		want := filepath.Join(dir, "shot.jpg")
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestPrintAndMaybeOpen(t *testing.T) {
	orig := openCmd
	t.Cleanup(func() { openCmd = orig })

	t.Run("no_open", func(t *testing.T) {
		var called []string
		openCmd = func(path string) error {
			called = append(called, path)
			return nil
		}
		var buf bytes.Buffer
		if err := printAndMaybeOpen(&buf, "/tmp/demo.png", false); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "/tmp/demo.png\n" {
			t.Errorf("stdout = %q, want path line", got)
		}
		if len(called) != 0 {
			t.Errorf("openCmd called with %v, want none", called)
		}
	})

	t.Run("with_open", func(t *testing.T) {
		var called []string
		openCmd = func(path string) error {
			called = append(called, path)
			return nil
		}
		var buf bytes.Buffer
		if err := printAndMaybeOpen(&buf, "/tmp/shot.png", true); err != nil {
			t.Fatal(err)
		}
		if got := buf.String(); got != "/tmp/shot.png\n" {
			t.Errorf("stdout = %q, want path line", got)
		}
		if len(called) != 1 || called[0] != "/tmp/shot.png" {
			t.Errorf("openCmd called with %v, want [/tmp/shot.png]", called)
		}
	})

	t.Run("open_fails", func(t *testing.T) {
		openCmd = func(path string) error {
			return errors.New("boom")
		}
		var buf bytes.Buffer
		err := printAndMaybeOpen(&buf, "/tmp/fail.png", true)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "open /tmp/fail.png") {
			t.Errorf("error = %v, want open path prefix", err)
		}
		// Path still printed before open failure.
		if got := buf.String(); got != "/tmp/fail.png\n" {
			t.Errorf("stdout = %q, want path printed before open error", got)
		}
	})
}
