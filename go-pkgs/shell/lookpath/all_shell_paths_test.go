package lookpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAllShellPATHsWithUnionFirstSeen(t *testing.T) {
	t.Parallel()
	sep := string(os.PathListSeparator)
	a := filepath.Join(t.TempDir(), "a")
	b := filepath.Join(t.TempDir(), "b")
	c := filepath.Join(t.TempDir(), "c")
	opts := Options{
		RunLogin: loginPATHByShell(map[string]string{
			"bash": strings.Join([]string{a, b}, sep),
			"zsh":  strings.Join([]string{b, c}, sep),
		}, nil),
	}
	got, err := AllShellPATHsWith(opts)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Clean(a), filepath.Clean(b), filepath.Clean(c)}
	assertStringSlice(t, "AllShellPATHs", got, want)
}

func TestAllShellPATHsWithBashFailZshHit(t *testing.T) {
	t.Parallel()
	sep := string(os.PathListSeparator)
	a := filepath.Join(t.TempDir(), "a")
	b := filepath.Join(t.TempDir(), "b")
	opts := Options{
		RunLogin: loginPATHByShell(map[string]string{
			"zsh": strings.Join([]string{a, b}, sep),
		}, map[string]error{
			"bash": fmt.Errorf("injected bash login failure"),
		}),
	}
	got, err := AllShellPATHsWith(opts)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Clean(a), filepath.Clean(b)}
	assertStringSlice(t, "AllShellPATHs", got, want)
}

func TestAllShellPATHsWithBothFail(t *testing.T) {
	t.Parallel()
	opts := Options{
		RunLogin: loginPATHByShell(nil, map[string]error{
			"bash": fmt.Errorf("bash fail"),
			"zsh":  fmt.Errorf("zsh fail"),
		}),
	}
	got, err := AllShellPATHsWith(opts)
	if err == nil {
		t.Fatalf("got %v, want error", got)
	}
	if got != nil {
		t.Fatalf("got paths %v, want nil on error", got)
	}
}

func TestAllShellPATHsWithBothEmpty(t *testing.T) {
	t.Parallel()
	opts := Options{
		RunLogin: loginPATHByShell(map[string]string{
			"bash": "",
			"zsh":  "",
		}, nil),
	}
	got, err := AllShellPATHsWith(opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %v, want empty", got)
	}
}

func TestLookupInAllShellPATHsEmptyName(t *testing.T) {
	t.Parallel()
	_, err := LookupInAllShellPATHs("", Options{
		RunLogin: loginPATHByShell(map[string]string{"bash": "/bin", "zsh": ""}, nil),
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestLookupInAllShellPATHsHitOneOfTwo(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := filepath.Join(root, "a")
	b := filepath.Join(root, "b")
	if err := os.MkdirAll(a, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(b, 0o755); err != nil {
		t.Fatal(err)
	}
	codexA := filepath.Join(a, "codex")
	if err := os.WriteFile(codexA, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	sep := string(os.PathListSeparator)
	opts := Options{
		RunLogin: loginPATHByShell(map[string]string{
			"bash": strings.Join([]string{a, b}, sep),
			"zsh":  "",
		}, nil),
	}
	got, err := LookupInAllShellPATHs("codex", opts)
	if err != nil {
		t.Fatal(err)
	}
	assertStringSlice(t, "LookupInAllShellPATHs", got, []string{filepath.Clean(codexA)})
}

func TestLookupInAllShellPATHsDedupeSameDir(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	a := filepath.Join(root, "a")
	if err := os.MkdirAll(a, 0o755); err != nil {
		t.Fatal(err)
	}
	codex := filepath.Join(a, "codex")
	if err := os.WriteFile(codex, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	opts := Options{
		RunLogin: loginPATHByShell(map[string]string{
			"bash": a,
			"zsh":  a,
		}, nil),
	}
	got, err := LookupInAllShellPATHs("codex", opts)
	if err != nil {
		t.Fatal(err)
	}
	assertStringSlice(t, "LookupInAllShellPATHs", got, []string{filepath.Clean(codex)})
}

func loginPATHByShell(paths map[string]string, fails map[string]error) func(shell, command string, env []string) (string, error) {
	return func(shell, command string, env []string) (string, error) {
		_ = command
		_ = env
		if err := fails[shell]; err != nil {
			return "", err
		}
		p := paths[shell]
		if p == "" {
			return "OTHER=1\x00", nil
		}
		return "PATH=" + p + "\x00", nil
	}
}

func assertStringSlice(t *testing.T, field string, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %#v, want %#v", field, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %#v, want %#v", field, got, want)
		}
	}
}
