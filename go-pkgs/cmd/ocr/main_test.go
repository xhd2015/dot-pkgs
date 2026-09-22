package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/ocr"
)

func TestRunMissingImage(t *testing.T) {
	var buf bytes.Buffer
	err := runWith(nil, &buf, ocr.Options{})
	if err == nil || !strings.Contains(err.Error(), "missing image path") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunExtraArgs(t *testing.T) {
	var buf bytes.Buffer
	err := runWith([]string{"a.png", "b.png"}, &buf, ocr.Options{})
	if err == nil || !strings.Contains(err.Error(), "unexpected argument") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunWithHelper(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "x.png")
	if err := os.WriteFile(img, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	helper := filepath.Join(dir, "helper")
	if err := os.WriteFile(helper, []byte("#!/bin/sh\necho CLI-OCR\n"), 0755); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err := runWith([]string{img}, &buf, ocr.Options{HelperPath: helper})
	if err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "CLI-OCR\n" {
		t.Fatalf("stdout = %q", got)
	}
}
