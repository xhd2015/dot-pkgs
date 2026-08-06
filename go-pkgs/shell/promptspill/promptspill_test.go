package promptspill_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/promptspill"
)

func TestMaybeSpill_UnderMax(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	full := "short 中文"
	short, path, err := promptspill.MaybeSpill(full, dir, promptspill.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if short != full || path != "" {
		t.Fatalf("want as-is, got short=%q path=%q", short, path)
	}
	matches, _ := filepath.Glob(filepath.Join(dir, "*-*.txt"))
	if len(matches) != 0 {
		t.Fatalf("unexpected files: %v", matches)
	}
}

func TestMaybeSpill_OverMaxRemainderOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	full := strings.Repeat("A", 868) + strings.Repeat("Z", 32)
	if len([]rune(full)) != 900 {
		t.Fatalf("fixture runes=%d", len([]rune(full)))
	}
	short, path, err := promptspill.MaybeSpill(full, dir, promptspill.Options{
		FilePrefix: "open-inject-full",
	})
	if err != nil {
		t.Fatal(err)
	}
	if path == "" {
		t.Fatal("empty path")
	}
	if len([]rune(short)) > promptspill.DefaultMaxRunes {
		t.Fatalf("short runes=%d", len([]rune(short)))
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) == full {
		t.Fatal("file is full copy")
	}
	idx := strings.LastIndex(short, promptspill.MarkerOpen)
	if idx < 0 {
		t.Fatalf("no marker: %q", short)
	}
	rest := short[idx+len(promptspill.MarkerOpen):]
	if !strings.HasSuffix(rest, promptspill.MarkerClose) {
		t.Fatal("bad marker close")
	}
	prefix := short[:idx]
	if prefix+string(body) != full {
		t.Fatal("reconstruct failed")
	}
	if !strings.HasPrefix(filepath.Base(path), "open-inject-full-") {
		t.Fatalf("basename %s", filepath.Base(path))
	}
	if !utf8.ValidString(short) {
		t.Fatal("invalid utf8 short")
	}
}

func TestMaybeSpill_ChineseRunes(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	full := strings.Repeat("中", 900)
	short, path, err := promptspill.MaybeSpill(full, dir, promptspill.Options{})
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	idx := strings.LastIndex(short, promptspill.MarkerOpen)
	prefix := short[:idx]
	if prefix+string(body) != full {
		t.Fatal("reconstruct failed")
	}
	if !utf8.ValidString(short) || !utf8.ValidString(string(body)) {
		t.Fatal("invalid utf8")
	}
}

func TestMaybeSpill_CollideSafe(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	full := strings.Repeat("B", 900)
	_, p1, err1 := promptspill.MaybeSpill(full, dir, promptspill.Options{FilePrefix: "open-inject-full"})
	_, p2, err2 := promptspill.MaybeSpill(full+"X", dir, promptspill.Options{FilePrefix: "open-inject-full"})
	if err1 != nil || err2 != nil {
		t.Fatalf("%v %v", err1, err2)
	}
	if p1 == p2 {
		t.Fatalf("same path %q", p1)
	}
}
