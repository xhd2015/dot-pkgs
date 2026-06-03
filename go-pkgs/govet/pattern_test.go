package govet

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func mkMod(t *testing.T, dir, module string) {
	t.Helper()
	content := fmt.Sprintf("module %s\n\ngo 1.24\n", module)
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
}

func mkGoFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func TestResolvePatterns_CurrentDir(t *testing.T) {
	dir := t.TempDir()
	mkMod(t, dir, "testmod")
	mkGoFile(t, dir, "main.go", `package main
import "fmt"
func main() { fmt.Println("hello") }
`)

	dirs, err := ResolvePatterns(dir, []string{"."})
	if err != nil {
		t.Fatalf("ResolvePatterns: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("expected 1 dir, got %d: %v", len(dirs), dirs)
	}
	if dirs[0] != dir {
		t.Errorf("expected %q, got %q", dir, dirs[0])
	}
}

func TestResolvePatterns_Recursive(t *testing.T) {
	root := t.TempDir()
	mkMod(t, root, "testmod")
	mkGoFile(t, root, "main.go", `package main`)
	mkGoFile(t, root, "sub/a/a.go", `package a`)
	mkGoFile(t, root, "sub/b/b.go", `package b`)
	mkGoFile(t, root, "sub/b/deep/c.go", `package deep`)

	// create empty dir (no Go files) — should be excluded
	if err := os.MkdirAll(filepath.Join(root, "sub", "empty"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	dirs, err := ResolvePatterns(root, []string{"./..."})
	if err != nil {
		t.Fatalf("ResolvePatterns: %v", err)
	}

	sort.Strings(dirs)
	expected := []string{
		root,
		filepath.Join(root, "sub", "a"),
		filepath.Join(root, "sub", "b"),
		filepath.Join(root, "sub", "b", "deep"),
	}
	if len(dirs) != len(expected) {
		t.Fatalf("expected %d dirs, got %d: %v", len(expected), len(dirs), dirs)
	}
	for i, d := range dirs {
		if d != expected[i] {
			t.Errorf("dirs[%d]: expected %q, got %q", i, expected[i], d)
		}
	}
}

func TestResolvePatterns_SubDirPattern(t *testing.T) {
	root := t.TempDir()
	mkMod(t, root, "testmod")
	mkGoFile(t, root, "main.go", `package main`)
	mkGoFile(t, root, "sub/a/a.go", `package a`)
	mkGoFile(t, root, "sub/b/b.go", `package b`)
	mkGoFile(t, root, "other/x.go", `package other`)

	dirs, err := ResolvePatterns(root, []string{"./sub/..."})
	if err != nil {
		t.Fatalf("ResolvePatterns: %v", err)
	}

	sort.Strings(dirs)
	expected := []string{
		filepath.Join(root, "sub", "a"),
		filepath.Join(root, "sub", "b"),
	}
	if len(dirs) != len(expected) {
		t.Fatalf("expected %d dirs, got %d: %v", len(expected), len(dirs), dirs)
	}
	for i, d := range dirs {
		if d != expected[i] {
			t.Errorf("dirs[%d]: expected %q, got %q", i, expected[i], d)
		}
	}
}

func TestResolvePatterns_NoGoFiles(t *testing.T) {
	root := t.TempDir()
	mkMod(t, root, "testmod")

	dirs, err := ResolvePatterns(root, []string{"./..."})
	if err != nil {
		t.Fatalf("ResolvePatterns: %v", err)
	}
	if len(dirs) != 0 {
		t.Errorf("expected 0 dirs for module with no Go files, got %d: %v", len(dirs), dirs)
	}
}

func TestResolvePatterns_EmptyPatterns(t *testing.T) {
	root := t.TempDir()
	mkMod(t, root, "testmod")
	mkGoFile(t, root, "main.go", `package main`)

	dirs, err := ResolvePatterns(root, nil)
	if err != nil {
		t.Fatalf("ResolvePatterns: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("expected 1 dir for empty patterns (defaults to .), got %d", len(dirs))
	}
	if dirs[0] != root {
		t.Errorf("expected %q, got %q", root, dirs[0])
	}
}

func TestResolvePatterns_TestFiles(t *testing.T) {
	root := t.TempDir()
	mkMod(t, root, "testmod")
	mkGoFile(t, root, "main.go", `package main`)
	mkGoFile(t, root, "main_test.go", `package main
import "testing"
func TestX(t *testing.T) {}
`)

	dirs, err := ResolvePatterns(root, []string{"./..."})
	if err != nil {
		t.Fatalf("ResolvePatterns: %v", err)
	}
	if len(dirs) != 1 {
		t.Fatalf("expected 1 dir (test files should be included), got %d: %v", len(dirs), dirs)
	}
}

func TestResolvePatterns_DotDotDotExcludesVendor(t *testing.T) {
	root := t.TempDir()
	mkMod(t, root, "testmod")
	mkGoFile(t, root, "main.go", `package main`)
	mkGoFile(t, root, "vendor/x/x.go", `package x`)

	dirs, err := ResolvePatterns(root, []string{"./..."})
	if err != nil {
		t.Fatalf("ResolvePatterns: %v", err)
	}
	for _, d := range dirs {
		if strings.Contains(d, "vendor") {
			t.Errorf("vendor directory should be excluded, got: %v", dirs)
		}
	}
}
