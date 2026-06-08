package gomodbadreplace

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChecker_NoReplace(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for go.mod without replace, got %d", len(v))
	}
}

func TestChecker_VersionReplace(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
replace example.com/old => example.com/old v1.2.3
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for version replace, got %d", len(v))
	}
}

func TestChecker_DifferentModuleReplace(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
replace example.com/old => example.com/new v1.0.0
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for different-module replace, got %d", len(v))
	}
}

func TestChecker_LocalRelativeReplace(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
replace example.com/old => ./local/path
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for local relative replace, got %d", len(v))
	}
	if !strings.Contains(v[0].Message, "./local/path") {
		t.Errorf("expected message to contain the local path, got %q", v[0].Message)
	}
}

func TestChecker_LocalParentReplace(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
replace example.com/old => ../sibling
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for local parent replace, got %d", len(v))
	}
}

func TestChecker_LocalAbsoluteReplace(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
replace example.com/old => /absolute/path
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for local absolute replace, got %d", len(v))
	}
}

func TestChecker_MixedReplaces(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
replace example.com/ok1 => example.com/ok1 v1.0.0
replace example.com/bad => ./local
replace example.com/ok2 => example.com/new v2.0.0
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for mixed replaces, got %d", len(v))
	}
}

func TestChecker_MultipleBadReplaces(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(writeGoMod(t, `
module example.com/m
go 1.24
replace example.com/a => ./a
replace example.com/b => ../b
`))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(v))
	}
}

func TestChecker_EmptyGoMod(t *testing.T) {
	c := &Checker{}
	path := writeGoMod(t, ``)
	v, err := c.Check(path)
	if err != nil {
		t.Logf("empty go.mod returned error (module parser behavior may vary): %v", err)
		return
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for empty go.mod, got %d", len(v))
	}
}

func TestChecker_NonExistentFile(t *testing.T) {
	c := &Checker{}
	_, err := c.Check(filepath.Join(t.TempDir(), "nonexistent.mod"))
	if err == nil {
		t.Fatal("expected error for non-existent file")
	}
}

func TestChecker_Name(t *testing.T) {
	c := &Checker{}
	if c.Name() != "gomod-bad-replace" {
		t.Errorf("expected name 'gomod-bad-replace', got %q", c.Name())
	}
}

func writeGoMod(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "go.mod")
	content = strings.TrimSpace(content)
	if err := os.WriteFile(path, []byte(content+"\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	return path
}
