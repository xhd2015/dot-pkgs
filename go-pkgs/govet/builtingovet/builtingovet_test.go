package builtingovet

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestChecker_CleanFile(t *testing.T) {
	dir := setupGoModDir(t, "package main\nfunc main() {}\n")
	c := &Checker{}
	v, err := c.Check([]string{filepath.Join(dir, "main.go")})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for clean file, got %d: %v", len(v), v)
	}
}

func TestChecker_VetIssue(t *testing.T) {
	dir := setupGoModDir(t, `package main
import "fmt"
func main() {
	fmt.Printf("%d", "string")
}
`)
	c := &Checker{}
	v, err := c.Check([]string{filepath.Join(dir, "main.go")})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) == 0 {
		t.Error("expected violations for printf format mismatch")
		return
	}
	found := false
	for _, violation := range v {
		if violation.Checker != "builtin-go-vet" {
			t.Errorf("expected checker 'builtin-go-vet', got %q", violation.Checker)
		}
		if strings.Contains(violation.Message, "Printf") || strings.Contains(violation.Message, "format") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected printf-related message, got %v", v)
	}
}

func TestChecker_VetUnreachableCode(t *testing.T) {
	dir := setupGoModDir(t, `package main
func main() {
	return
	println("unreachable")
}
`)
	c := &Checker{}
	v, err := c.Check([]string{filepath.Join(dir, "main.go")})
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) == 0 {
		t.Error("expected violations for unreachable code")
		return
	}
	found := false
	for _, violation := range v {
		if strings.Contains(violation.Message, "unreachable") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected unreachable code message, got %v", v)
	}
}

func TestChecker_EmptyFileList(t *testing.T) {
	c := &Checker{}
	v, err := c.Check(nil)
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for empty file list, got %d", len(v))
	}
}

func TestChecker_NoGoMod(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, dir, "main.go", "package main\nfunc main() {}\n")
	c := &Checker{}
	_, err := c.Check([]string{filepath.Join(dir, "main.go")})
	if err != nil {
		t.Logf("go vet without go.mod returned error (expected on some Go versions): %v", err)
	}
}

func TestChecker_Name(t *testing.T) {
	c := &Checker{}
	if c.Name() != "builtin-go-vet" {
		t.Errorf("expected name 'builtin-go-vet', got %q", c.Name())
	}
}

func setupGoModDir(t *testing.T, mainGoContent string) string {
	t.Helper()
	dir := t.TempDir()
	mustWriteFile(t, dir, "go.mod", "module testmod\ngo 1.24\n")
	mustWriteFile(t, dir, "main.go", mainGoContent)
	return dir
}

func mustWriteFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}
