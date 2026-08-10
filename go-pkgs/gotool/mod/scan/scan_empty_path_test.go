package scan_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/gotool/mod/scan"
)

// go.mod with no module line is a parent-module boundary only; Scan must not
// emit a Module for it (empty Path is not a publishable module).
func TestScanSkipsEmptyModulePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	boundary := filepath.Join(dir, "boundary")
	if err := os.MkdirAll(boundary, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(boundary, "go.mod"), []byte("// Not a Go module.\n// Boundary only.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	mods, err := scan.Scan(dir, scan.Options{})
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if len(mods) != 1 {
		t.Fatalf("got %d modules, want 1: %+v", len(mods), mods)
	}
	if mods[0].Dir != "." || mods[0].Path != "example.com/root" {
		t.Fatalf("root = {Dir:%q Path:%q}, want {. example.com/root}", mods[0].Dir, mods[0].Path)
	}
}
