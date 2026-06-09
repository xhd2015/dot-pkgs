package files

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestResolveGoFiles_EmptyPaths(t *testing.T) {
	files, err := ResolveGoFiles("/tmp", nil)
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestResolveGoFiles_Dir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", "package main\n")
	writeFile(t, dir, "b.go", "package main\n")
	writeFile(t, dir, "README.md", "# readme\n")

	files, err := ResolveGoFiles("", []string{dir})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 .go files, got %d: %v", len(files), files)
	}
}

func TestResolveGoFiles_SingleGoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, dir, "main.go", "package main\n")

	files, err := ResolveGoFiles(dir, []string{path})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file, got %d", len(files))
	}
	if files[0] != path {
		t.Errorf("expected %s, got %s", path, files[0])
	}
}

func TestResolveGoFiles_NonGoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "README.md")
	writeFile(t, dir, "README.md", "# readme\n")

	files, err := ResolveGoFiles(dir, []string{path})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files for non-go, got %d", len(files))
	}
}

func TestResolveGoFiles_VendorWithGoMod(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module testmod\n")
	writeFile(t, dir, "main.go", "package main\n")
	vendorDir := filepath.Join(dir, "vendor")
	mustMkdirAll(t, vendorDir)
	writeFile(t, vendorDir, "thirdparty.go", "package thirdparty\n")

	files, err := ResolveGoFiles("", []string{dir})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (vendor skipped), got %d: %v", len(files), files)
	}
	if filepath.Base(files[0]) != "main.go" {
		t.Errorf("expected main.go, got %s", files[0])
	}
}

func TestResolveGoFiles_VendorWithoutGoMod(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	vendorDir := filepath.Join(dir, "vendor")
	mustMkdirAll(t, vendorDir)
	writeFile(t, vendorDir, "thirdparty.go", "package thirdparty\n")

	files, err := ResolveGoFiles("", []string{dir})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 2 {
		t.Errorf("expected 2 files (vendor not skipped), got %d: %v", len(files), files)
	}
}

func TestResolveGoFiles_Deduplicate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, dir, "main.go", "package main\n")

	files, err := ResolveGoFiles(dir, []string{path, path})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file after dedup, got %d: %v", len(files), files)
	}
}

func TestResolveGoFiles_DirRecursive(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	subDir := filepath.Join(dir, "sub")
	mustMkdirAll(t, subDir)
	writeFile(t, subDir, "sub.go", "package sub\n")

	files, err := ResolveGoFiles("", []string{dir})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	sort.Strings(files)
	if len(files) != 2 {
		t.Errorf("expected 2 files, got %d: %v", len(files), files)
	}
}

func TestResolveGoFiles_SkipDotGit(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", "package main\n")
	gitDir := filepath.Join(dir, ".git")
	mustMkdirAll(t, gitDir)
	writeFile(t, gitDir, "config.go", "package git\n")

	files, err := ResolveGoFiles("", []string{dir})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("expected 1 file (.git skipped), got %d: %v", len(files), files)
	}
}

func TestResolveGoFiles_VendorFile(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module testmod\n")
	vendorDir := filepath.Join(dir, "vendor")
	mustMkdirAll(t, vendorDir)
	vendorFile := filepath.Join(vendorDir, "thirdparty.go")
	writeFile(t, vendorDir, "thirdparty.go", "package thirdparty\n")

	files, err := ResolveGoFiles(dir, []string{vendorFile})
	if err != nil {
		t.Fatalf("ResolveGoFiles: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("expected 0 files (vendor file filtered), got %d: %v", len(files), files)
	}
}

func TestResolveGoFiles_NonExistentPath(t *testing.T) {
	_, err := ResolveGoFiles("/tmp", []string{"/nonexistent/path"})
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}
