package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenLogCreatesParentAndAppends(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "svc.log")
	if err := AppendLog(path, "first %d", 1); err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}
	if err := AppendLog(path, "second %d", 2); err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := strings.Count(string(data), "\n"); got != 2 {
		t.Fatalf("log = %q, want two appended lines", string(data))
	}
	if !strings.Contains(string(data), "first 1") || !strings.Contains(string(data), "second 2") {
		t.Fatalf("log = %q, want both entries", string(data))
	}
}

func TestAppendLogDoesNotDoubleNewlines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "svc.log")
	if err := AppendLog(path, "already terminated\n"); err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}
	if err := AppendLog(path, "\n"); err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if got := string(data); got != "already terminated\n\n" {
		t.Fatalf("log = %q, want no doubled newline", got)
	}
}

func TestTailLogReturnsLastLinesOldestFirst(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns no processes but writes a file")
	}
	path := filepath.Join(t.TempDir(), "svc.log")
	var b strings.Builder
	for i := 1; i <= 20; i++ {
		b.WriteString("line-")
		b.WriteString(strings.Repeat("x", i))
		b.WriteString("\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	lines, err := TailLog(path, 3)
	if err != nil {
		t.Fatalf("TailLog() error = %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("TailLog() = %v, want 3 lines", lines)
	}
	if !strings.Contains(lines[0], "xxxxxxxxxxxxxxxx") || !strings.Contains(lines[2], strings.Repeat("x", 20)) {
		t.Fatalf("TailLog() = %v, want the last three lines oldest first", lines)
	}
}

func TestTailLogMissingFile(t *testing.T) {
	lines, err := TailLog(filepath.Join(t.TempDir(), "absent.log"), 5)
	if err != nil {
		t.Fatalf("TailLog() error = %v, want nil for a missing file", err)
	}
	if len(lines) != 0 {
		t.Fatalf("TailLog() = %v, want no lines", lines)
	}
}

func TestTailLogNonPositiveLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "svc.log")
	if err := os.WriteFile(path, []byte("one\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	for _, n := range []int{0, -1} {
		lines, err := TailLog(path, n)
		if err != nil || lines != nil {
			t.Fatalf("TailLog(%d) = %v, %v, want nil, nil", n, lines, err)
		}
	}
}

func TestTailLogKeepsPartialLastLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "svc.log")
	if err := os.WriteFile(path, []byte("complete\nin progress"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	lines, err := TailLog(path, 2)
	if err != nil {
		t.Fatalf("TailLog() error = %v", err)
	}
	if len(lines) != 2 || lines[1] != "in progress" {
		t.Fatalf("TailLog() = %v, want the unterminated last line", lines)
	}
}
