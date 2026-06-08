package govet

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations in empty dir, got %d: %v", len(v), v)
	}
}

func TestRun_CleanPackage(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`)
	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for clean package, got %d: %v", len(v), v)
	}
}

func TestRun_FileLengthViolation(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", lines(600, "// comment\n"))

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Checker != "file-length" {
		t.Errorf("expected file-length checker, got %q", v[0].Checker)
	}
}

func TestRun_FileLengthUnderThreshold(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", lines(100, "// comment\n"))

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected no violations for file under threshold, got %d", len(v))
	}
}

func TestRun_StdFlagViolation(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	hasFlag := false
	for _, violation := range v {
		if violation.Checker == "std-flag" {
			hasFlag = true
		}
	}
	if !hasFlag {
		t.Errorf("expected std-flag violation, got violations: %v", v)
	}
}

func TestRun_BothViolations(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", `package main
import "flag"
`+lines(600, "// comment\n"))

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	hasFileLen := false
	hasStdFlag := false
	for _, violation := range v {
		switch violation.Checker {
		case "file-length":
			hasFileLen = true
		case "std-flag":
			hasStdFlag = true
		}
	}
	if !hasFileLen {
		t.Error("missing file-length violation")
	}
	if !hasStdFlag {
		t.Error("missing std-flag violation")
	}
}

func TestRun_ExcludesChecker(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "main.go", `package main
import "flag"
`+lines(600, "// comment\n"))

	v, err := Run(Config{
		FileMaxLines: 500,
		Excludes:     []string{"file-length"},
		Paths:        []string{dir},
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, violation := range v {
		if violation.Checker == "file-length" {
			t.Errorf("file-length violation should be excluded: %v", violation)
		}
	}
}

func TestRun_NonExistentDir(t *testing.T) {
	_, err := Run(Config{FileMaxLines: 500, Paths: []string{"/nonexistent/path/xyz"}})
	if err == nil {
		t.Fatal("expected error for non-existent dir")
	}
}

func TestRun_MultipleDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	writeFile(t, dir1, "a.go", `package main
import "flag"
`)
	writeFile(t, dir2, "b.go", `package main
import "flag"
`)

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir1, dir2}})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	count := 0
	for _, violation := range v {
		if violation.Checker == "std-flag" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 std-flag violations, got %d", count)
	}
}

func TestRun_PathsSingleViolatingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{path}} )
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	hasFlag := false
	for _, violation := range v {
		if violation.Checker == "std-flag" {
			hasFlag = true
		}
	}
	if !hasFlag {
		t.Errorf("expected std-flag violation, got %v", v)
	}
}

func TestRun_PathsSingleCleanFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "main.go")
	writeFile(t, dir, "main.go", `package main
import "fmt"
func main() { fmt.Println("hello") }
`)

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{path}} )
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected 0 violations, got %d: %v", len(v), v)
	}
}

func TestRun_PathsTwoFilesOneViolation(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.go")
	path2 := filepath.Join(dir, "b.go")
	writeFile(t, dir, "a.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	writeFile(t, dir, "b.go", `package main
import "fmt"
func main() { fmt.Println("hello") }
`)

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{path1, path2}} )
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	count := 0
	for _, violation := range v {
		if violation.Checker == "std-flag" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected 1 std-flag violation, got %d", count)
	}
}

func TestRun_PathsDir(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	writeFile(t, dir, "b.go", `package main
import "flag"
func main() { flag.Parse() }
`)

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}} )
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	count := 0
	for _, violation := range v {
		if violation.Checker == "std-flag" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 std-flag violations, got %d", count)
	}
}

func TestRun_PathsDirRecursive(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "sub")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, dir, "main.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	if err := os.WriteFile(filepath.Join(subDir, "sub.go"), []byte(`package main
import "flag"
func main() { flag.Parse() }
`), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{dir}} )
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	count := 0
	for _, violation := range v {
		if violation.Checker == "std-flag" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 violations (dir is recursive), got %d", count)
	}
}

func TestRun_PathsNonGoFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "README.md")
	if err := os.WriteFile(path, []byte("# hello"), 0644); err != nil {
		t.Fatalf("writeFile: %v", err)
	}

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{path}} )
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(v) != 0 {
		t.Errorf("expected 0 violations for non-go file, got %d", len(v))
	}
}

func TestRun_PathsMixFileAndDir(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()
	path1 := filepath.Join(dir1, "a.go")
	writeFile(t, dir1, "a.go", `package main
import "flag"
func main() { flag.Parse() }
`)
	writeFile(t, dir2, "b.go", `package main
import "flag"
func main() { flag.Parse() }
`)

	v, err := Run(Config{FileMaxLines: 500, Paths: []string{path1, dir2}} )
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	count := 0
	for _, violation := range v {
		if violation.Checker == "std-flag" {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 std-flag violations, got %d", count)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

func lines(n int, line string) string {
	return strings.Repeat(line, n)
}
