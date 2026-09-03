package logs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScanLinesFromTail_newestFirst(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.log")
	if err := os.WriteFile(path, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var got []string
	err := ScanLinesFromTail(path, ScanLinesFromTailOptions{ChunkSize: 2}, func(line string) (bool, error) {
		got = append(got, line)
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"c", "b", "a"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestScanLinesFromTail_stopEarly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.log")
	if err := os.WriteFile(path, []byte("a\nb\nc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var got []string
	err := ScanLinesFromTail(path, ScanLinesFromTailOptions{}, func(line string) (bool, error) {
		got = append(got, line)
		return true, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "c" {
		t.Fatalf("got %v want [c]", got)
	}
}

func TestScanLinesFromTail_noTrailingNewline(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.log")
	if err := os.WriteFile(path, []byte("a\nb"), 0o644); err != nil {
		t.Fatal(err)
	}
	var got []string
	err := ScanLinesFromTail(path, ScanLinesFromTailOptions{ChunkSize: 1}, func(line string) (bool, error) {
		got = append(got, line)
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"b", "a"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestScanLinesFromTail_growPastLongLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.log")
	long := strings.Repeat("x", 100)
	if err := os.WriteFile(path, []byte(long+"\nend\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var got []string
	err := ScanLinesFromTail(path, ScanLinesFromTailOptions{ChunkSize: 8}, func(line string) (bool, error) {
		got = append(got, line)
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "end" || got[1] != long {
		t.Fatalf("got %v", got)
	}
}

func TestScanLinesFromTail_empty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.log")
	if err := os.WriteFile(path, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	n := 0
	err := ScanLinesFromTail(path, ScanLinesFromTailOptions{}, func(string) (bool, error) {
		n++
		return false, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("calls=%d", n)
	}
}
