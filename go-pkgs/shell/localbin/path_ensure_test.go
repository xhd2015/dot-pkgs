package localbin

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyChecker_AppendEmpty(t *testing.T) {
	got, action, orphan := ApplyChecker("", CheckerBlock())
	if action != "appended" || orphan {
		t.Fatalf("action=%q orphan=%v", action, orphan)
	}
	if got != CheckerBlock() {
		t.Fatalf("got:\n%s\nwant:\n%s", got, CheckerBlock())
	}
}

func TestApplyChecker_Unchanged(t *testing.T) {
	prefix := "export FOO=1\n"
	content := prefix + CheckerBlock()
	got, action, orphan := ApplyChecker(content, CheckerBlock())
	if action != "unchanged" || orphan {
		t.Fatalf("action=%q orphan=%v", action, orphan)
	}
	if got != content {
		t.Fatalf("should keep original bytes")
	}
}

func TestApplyChecker_ReplaceDifferent(t *testing.T) {
	old := checkerBegin + "\nexport PATH=/old\n" + checkerEnd + "\n"
	content := "# keep\n" + old + "# after\n"
	got, action, orphan := ApplyChecker(content, CheckerBlock())
	if action != "replaced" || orphan {
		t.Fatalf("action=%q orphan=%v", action, orphan)
	}
	want := "# keep\n" + CheckerBlock() + "# after\n"
	if got != want {
		t.Fatalf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestApplyChecker_CollapseDuplicates(t *testing.T) {
	block := CheckerBlock()
	content := block + "\nmiddle\n" + block
	got, action, orphan := ApplyChecker(content, block)
	if action != "replaced" || orphan {
		t.Fatalf("action=%q orphan=%v", action, orphan)
	}
	if strings.Count(got, checkerBegin) != 1 {
		t.Fatalf("want one block, got:\n%s", got)
	}
	if !strings.Contains(got, "middle") {
		t.Fatalf("lost surrounding text:\n%s", got)
	}
}

func TestApplyChecker_OrphanBeginAppends(t *testing.T) {
	content := checkerBegin + "\nexport PATH=/partial\n"
	got, action, orphan := ApplyChecker(content, CheckerBlock())
	if action != "appended" || !orphan {
		t.Fatalf("action=%q orphan=%v", action, orphan)
	}
	if !strings.Contains(got, "export PATH=/partial") {
		t.Fatalf("orphan body should stay:\n%s", got)
	}
	if strings.Count(got, checkerBegin) != 2 {
		t.Fatalf("want orphan + new block, got:\n%s", got)
	}
}

func TestEnsureCheckerInFile_CreateAndIdempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".zshrc")
	action, err := EnsureCheckerInFile(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if action != "created" {
		t.Fatalf("action=%q want created", action)
	}
	st1, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	action, err = EnsureCheckerInFile(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if action != "unchanged" {
		t.Fatalf("action=%q want unchanged", action)
	}
	st2, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !st1.ModTime().Equal(st2.ModTime()) {
		t.Fatal("identical content should not rewrite the file")
	}
}

func TestEnsureCheckerInFile_SkipMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".profile")
	action, err := EnsureCheckerInFile(path, false)
	if err != nil {
		t.Fatal(err)
	}
	if action != "skipped_missing" {
		t.Fatalf("action=%q want skipped_missing", action)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("must not create exist-only rc file")
	}
}

func TestEnsureOnPATH_CreatesAndExistOnly(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".profile"), []byte("export FOO=1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	var stderr bytes.Buffer
	res := EnsureOnPATH(EnsureOpts{Home: home, DestDir: dest, Stderr: &stderr})
	if res.Skipped {
		t.Fatal("should not skip default dest")
	}
	if len(res.Updated) == 0 {
		t.Fatalf("expected updates, stderr=%q", stderr.String())
	}

	for _, name := range []string{".bash_profile", ".bashrc", ".zshrc", ".profile"} {
		body, err := os.ReadFile(filepath.Join(home, name))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if !strings.Contains(string(body), checkerBegin) {
			t.Fatalf("%s missing checker:\n%s", name, body)
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".zprofile")); !os.IsNotExist(err) {
		t.Fatal(".zprofile must not be created")
	}
	stderr.Reset()
	res = EnsureOnPATH(EnsureOpts{Home: home, DestDir: dest, Stderr: &stderr})
	if len(res.Updated) != 0 {
		t.Fatalf("second run should be idempotent, updated=%v stderr=%q", res.Updated, stderr.String())
	}
	if !strings.Contains(stderr.String(), "PATH already includes") {
		t.Fatalf("stderr=%q", stderr.String())
	}
	body, err := os.ReadFile(filepath.Join(home, ".profile"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(body), checkerBegin) != 1 {
		t.Fatalf("duplicate checker in .profile:\n%s", body)
	}
	if !strings.Contains(string(body), "export FOO=1") {
		t.Fatal("lost existing .profile content")
	}
}

func TestEnsureOnPATH_SkipsOtherDest(t *testing.T) {
	home := t.TempDir()
	var stderr bytes.Buffer
	res := EnsureOnPATH(EnsureOpts{Home: home, DestDir: "/usr/local/bin", Stderr: &stderr})
	if !res.Skipped {
		t.Fatal("want skipped for non-default dest")
	}
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); !os.IsNotExist(err) {
		t.Fatal("must not touch rc files when dest is not ~/.local/bin")
	}
}

func TestEnsureOnPATH_DryRun_NoWrite(t *testing.T) {
	home := t.TempDir()
	dest := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	res := EnsureOnPATH(EnsureOpts{
		Home:    home,
		DestDir: dest,
		DryRun:  true,
		Stdout:  &stdout,
		Stderr:  &stderr,
	})
	if res.Skipped {
		t.Fatal("should not skip default dest")
	}
	if len(res.Updated) == 0 {
		t.Fatalf("expected would-mutate plans, stdout=%q", stdout.String())
	}
	out := stdout.String()
	if !strings.Contains(out, "[dry-run] would create ~/.zshrc (PATH checker)") {
		t.Fatalf("stdout=%q", out)
	}
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not create .zshrc")
	}
	// After live ensure, dry-run should skip.
	EnsureOnPATH(EnsureOpts{Home: home, DestDir: dest, Stderr: io.Discard})
	stdout.Reset()
	res = EnsureOnPATH(EnsureOpts{
		Home:    home,
		DestDir: dest,
		DryRun:  true,
		Stdout:  &stdout,
		Stderr:  &stderr,
	})
	if len(res.Updated) != 0 {
		t.Fatalf("expected no mutations planned, updated=%v stdout=%q", res.Updated, stdout.String())
	}
	if !strings.Contains(stdout.String(), "[dry-run] skip: ~/.zshrc (PATH checker already present)") {
		t.Fatalf("stdout=%q", stdout.String())
	}
}
