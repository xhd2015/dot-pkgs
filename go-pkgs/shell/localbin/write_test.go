package localbin

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForwarderBody(t *testing.T) {
	got := ForwarderBody([]string{"tsk", "project", "add"})
	want := "#!/bin/sh\nexec tsk project add \"$@\"\n"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestForwarderBody_QuotesUnsafe(t *testing.T) {
	got := ForwarderBody([]string{"my tool", "a'b"})
	if !strings.HasPrefix(got, "#!/bin/sh\n") {
		t.Fatalf("missing shebang: %q", got)
	}
	if !strings.Contains(got, "'my tool'") {
		t.Fatalf("expected quoted arg: %q", got)
	}
	if !strings.Contains(got, `'a'\''b'`) {
		t.Fatalf("expected escaped quote: %q", got)
	}
}

func TestWriteScript(t *testing.T) {
	dir := t.TempDir()
	path, err := WriteScript(dir, "pmark", ForwarderBody([]string{"tsk", "project", "add"}))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "pmark")
	if path != want {
		t.Fatalf("path=%q want %q", path, want)
	}
	st, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode().Perm()&0o111 == 0 {
		t.Fatalf("not executable: mode=%v", st.Mode())
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != ForwarderBody([]string{"tsk", "project", "add"}) {
		t.Fatalf("body=%q", body)
	}
	// overwrite
	path2, err := WriteScript(dir, "pmark", "#!/bin/sh\necho hi\n")
	if err != nil {
		t.Fatal(err)
	}
	if path2 != path {
		t.Fatalf("path changed: %q", path2)
	}
	body, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "#!/bin/sh\necho hi\n" {
		t.Fatalf("overwrite body=%q", body)
	}
}

func TestWriteScript_RejectsBadName(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteScript(dir, "../x", "x\n"); err == nil {
		t.Fatal("expected error for path separator in name")
	}
	if _, err := WriteScript(dir, "", "x\n"); err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDefaultDir(t *testing.T) {
	got, err := DefaultDir("/tmp/home")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/home", ".local", "bin")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if !IsDefaultDest(want, "/tmp/home") {
		t.Fatal("IsDefaultDest should be true")
	}
	if IsDefaultDest("/usr/local/bin", "/tmp/home") {
		t.Fatal("IsDefaultDest should be false for /usr/local/bin")
	}
}
