package govet

import (
	"strings"
	"testing"
)

func TestFileLenChecker_Empty(t *testing.T) {
	c := &FileLenChecker{MaxLines: 500}
	v := c.CheckFile("foo.go", []byte{})
	if len(v) != 0 {
		t.Errorf("expected no violations for empty file, got %d", len(v))
	}
}

func TestFileLenChecker_OneLineWithNewline(t *testing.T) {
	c := &FileLenChecker{MaxLines: 500}
	v := c.CheckFile("foo.go", []byte("package main\n"))
	if len(v) != 0 {
		t.Errorf("expected no violations for 1-line file, got %d", len(v))
	}
}

func TestFileLenChecker_OneLineWithoutNewline(t *testing.T) {
	c := &FileLenChecker{MaxLines: 500}
	v := c.CheckFile("foo.go", []byte("package main"))
	if len(v) != 0 {
		t.Errorf("expected no violations for 1-line file without newline, got %d", len(v))
	}
}

func TestFileLenChecker_AtThreshold(t *testing.T) {
	c := &FileLenChecker{MaxLines: 500}
	lines := strings.Repeat("x\n", 500)
	v := c.CheckFile("foo.go", []byte(lines))
	if len(v) != 0 {
		t.Errorf("expected no violations at threshold (500), got %d: %v", len(v), v)
	}
}

func TestFileLenChecker_ExceedsThreshold(t *testing.T) {
	c := &FileLenChecker{MaxLines: 500}
	lines := strings.Repeat("x\n", 501)
	v := c.CheckFile("foo.go", []byte(lines))
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Checker != "file-length" {
		t.Errorf("expected checker name 'file-length', got %q", v[0].Checker)
	}
	if v[0].File != "foo.go" {
		t.Errorf("expected file 'foo.go', got %q", v[0].File)
	}
	if !strings.Contains(v[0].Message, "501") {
		t.Errorf("expected message to contain line count 501, got %q", v[0].Message)
	}
	if !strings.Contains(v[0].Message, "500") {
		t.Errorf("expected message to contain max 500, got %q", v[0].Message)
	}
}

func TestFileLenChecker_ExceedsNoTrailingNewline(t *testing.T) {
	c := &FileLenChecker{MaxLines: 500}
	lines := strings.Repeat("x\n", 500) + "x"
	v := c.CheckFile("foo.go", []byte(lines))
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if !strings.Contains(v[0].Message, "501") {
		t.Errorf("expected message to contain line count 501, got %q", v[0].Message)
	}
}

func TestFileLenChecker_MaxLinesZero(t *testing.T) {
	c := &FileLenChecker{MaxLines: 0}
	lines := strings.Repeat("x\n", 1000)
	v := c.CheckFile("foo.go", []byte(lines))
	if len(v) != 0 {
		t.Errorf("expected no violations when MaxLines is 0 (disabled), got %d", len(v))
	}
}

func TestFileLenChecker_Name(t *testing.T) {
	c := &FileLenChecker{MaxLines: 500}
	if c.Name() != "file-length" {
		t.Errorf("expected name 'file-length', got %q", c.Name())
	}
}
