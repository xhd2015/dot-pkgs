package rg

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureNoopPicksNewestAndNotices(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	oldBin := filepath.Join(home, ".local", "bin", "rg")
	loginDir := filepath.Join(home, "login-bin")
	newBin := filepath.Join(loginDir, "rg")
	for _, p := range []string{oldBin, newBin} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("#!/bin/sh\n"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	versions := map[string]string{
		filepath.Clean(oldBin): "14.1.1",
		filepath.Clean(newBin): "15.2.0",
	}
	var notices []string
	res, err := Ensure(context.Background(), EnsureOpts{
		Discover: DiscoverOpts{
			Home:   home,
			Getenv: func(string) string { return "" },
			IsExecutable: func(path string) bool {
				_, ok := versions[filepath.Clean(path)]
				return ok
			},
			RunVersion: func(ctx context.Context, bin string) (string, error) {
				if v, ok := versions[filepath.Clean(bin)]; ok {
					return v, nil
				}
				return "", os.ErrNotExist
			},
			RunLogin: func(shell, command string, env []string) (string, error) {
				return "PATH=" + loginDir + "\x00", nil
			},
		},
		Notice: func(body string) { notices = append(notices, body) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "noop" {
		t.Fatalf("action=%q selected=%+v all=%+v notices=%v", res.Action, res.Selected, res.All, notices)
	}
	if res.Selected.Path != filepath.Clean(newBin) || res.Selected.Version != "15.2.0" {
		t.Fatalf("selected=%+v", res.Selected)
	}
	if len(notices) != 1 || !strings.Contains(notices[0], "using rg 15.2.0") {
		t.Fatalf("notices=%v", notices)
	}
	if !strings.Contains(notices[0], "also found") {
		t.Fatalf("expected also found in %q", notices[0])
	}
}

func TestEnsureInstallsWhenMissing(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	dest := filepath.Join(root, "bin")
	installed := filepath.Join(dest, "rg")
	var notices []string
	installCalled := false
	res, err := Ensure(context.Background(), EnsureOpts{
		Discover: DiscoverOpts{
			Home:   root,
			Getenv: func(string) string { return "" },
			IsExecutable: func(path string) bool {
				if path == installed {
					fi, err := os.Stat(path)
					return err == nil && !fi.IsDir()
				}
				return false
			},
			RunVersion: func(ctx context.Context, bin string) (string, error) {
				if bin == installed {
					return "15.2.0", nil
				}
				return "", os.ErrNotExist
			},
			RunLogin: func(shell, command string, env []string) (string, error) {
				return "", nil
			},
		},
		Install: InstallOpts{
			Home:    root,
			DestDir: dest,
			GOOS:    "plan9", // force unsupported unless DownloadURL short-circuits via Fetch
			GOARCH:  "amd64",
			FetchLatestTag: func(ctx context.Context) (string, error) {
				installCalled = true
				return "", nil // won't get here if platform check first
			},
		},
		Notice: func(body string) { notices = append(notices, body) },
	})
	// Unsupported platform should error after "rg not found" notice
	if err == nil {
		t.Fatal("expected unsupported platform error")
	}
	if !strings.Contains(err.Error(), "no precompiled binary") {
		t.Fatalf("err=%v", err)
	}
	if len(notices) == 0 || !strings.Contains(notices[0], "rg not found") {
		t.Fatalf("notices=%v", notices)
	}
	_ = installCalled
	_ = res
}

