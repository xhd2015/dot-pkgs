package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/getclipboard"
)

func TestOCRImageClipboard(t *testing.T) {
	orig := ocrImage
	t.Cleanup(func() { ocrImage = orig })

	png := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0x00}
	ocrImage = func(ctx context.Context, data []byte, ext string) (string, error) {
		if ext != "png" {
			t.Fatalf("ext = %q", ext)
		}
		if !bytes.Equal(data, png) {
			t.Fatalf("unexpected image bytes")
		}
		return "recognized text", nil
	}

	var buf bytes.Buffer
	err := runWithOutput([]string{"--ocr"}, &buf, &fakeClipboard{image: png})
	if err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "recognized text\n" {
		t.Fatalf("stdout = %q", got)
	}
}

func TestOCRRequiresImage(t *testing.T) {
	var buf bytes.Buffer
	err := runWithOutput([]string{"--ocr"}, &buf, &fakeClipboard{
		formats: []getclipboard.Format{{
			MIME: "application/pdf",
			Read: func() []byte { return []byte("%PDF-1.4") },
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "--ocr requires image") {
		t.Fatalf("err = %v", err)
	}
}

func TestOCRMutuallyExclusiveWithOutput(t *testing.T) {
	var buf bytes.Buffer
	err := runWithOutput([]string{"--ocr", "-o", "/tmp/x"}, &buf, &fakeClipboard{})
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("err = %v", err)
	}
}

func TestOCRIgnoredForText(t *testing.T) {
	orig := ocrImage
	t.Cleanup(func() { ocrImage = orig })
	called := false
	ocrImage = func(ctx context.Context, data []byte, ext string) (string, error) {
		called = true
		return "", nil
	}
	var buf bytes.Buffer
	err := runWithOutput([]string{"--ocr"}, &buf, &fakeClipboard{text: []byte("plain")})
	if err != nil {
		t.Fatal(err)
	}
	if called {
		t.Fatal("ocrImage should not run for text")
	}
	if got := buf.String(); got != "plain" {
		t.Fatalf("stdout = %q", got)
	}
}

type fakeClipboard struct {
	image   []byte
	text    []byte
	formats []getclipboard.Format
}

func (f *fakeClipboard) Init() error            { return nil }
func (f *fakeClipboard) ReadImage() []byte      { return f.image }
func (f *fakeClipboard) ReadText() []byte       { return f.text }
func (f *fakeClipboard) ExtraFormats() []getclipboard.Format {
	return f.formats
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
		if got := buf.String(); got != "/tmp/fail.png\n" {
			t.Errorf("stdout = %q, want path printed before open error", got)
		}
	})
}
