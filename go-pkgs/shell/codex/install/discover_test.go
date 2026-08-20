package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWellKnownPathsWithHome(t *testing.T) {
	t.Parallel()
	home := filepath.Join(t.TempDir(), "home")
	got := WellKnownPaths(WellKnownPathsOpts{Home: home})
	want := []string{
		filepath.Join(home, ".local", "bin", "codex"),
		filepath.Join(home, "go", "bin", "codex"),
		"/opt/homebrew/bin/codex",
		"/usr/local/bin/codex",
		filepath.Join(home, ".npm-global", "bin", "codex"),
		filepath.Join(home, ".volta", "bin", "codex"),
	}
	assertStringSlice(t, "WellKnownPaths", got, want)
}

func TestWellKnownPathsEmptyHomeUsesUserHome(t *testing.T) {
	t.Parallel()
	got := WellKnownPaths(WellKnownPathsOpts{})
	if len(got) < 2 {
		t.Fatalf("WellKnownPaths empty home = %#v, want system + optional home entries", got)
	}
	if got[len(got)-4] != "/opt/homebrew/bin/codex" && got[0] != "/opt/homebrew/bin/codex" {
		// Either home-relative first then system, or system-only if UserHomeDir failed.
		found := false
		for _, p := range got {
			if p == "/opt/homebrew/bin/codex" {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("WellKnownPaths missing /opt/homebrew/bin/codex: %#v", got)
		}
	}
}

func TestVersionParsesCodexCLIPrefix(t *testing.T) {
	t.Parallel()
	bin := writeVersionScript(t, t.TempDir(), "codex-cli 0.147.0")
	got, err := Version(context.Background(), bin)
	if err != nil {
		t.Fatal(err)
	}
	if got != "0.147.0" {
		t.Fatalf("Version = %q, want 0.147.0", got)
	}
}

func TestVersionEmptyBinError(t *testing.T) {
	t.Parallel()
	_, err := Version(context.Background(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestVersionGarbageError(t *testing.T) {
	t.Parallel()
	bin := writeVersionScript(t, t.TempDir(), "not-a-version")
	_, err := Version(context.Background(), bin)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFoundCodexWellKnownAndLogin(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	local := filepath.Join(home, ".local", "bin", "codex")
	loginDir := filepath.Join(home, "login-bin")
	loginBin := filepath.Join(loginDir, "codex")
	versions := map[string]string{
		filepath.Clean(local):    "codex-cli 0.1.0",
		filepath.Clean(loginBin): "codex-cli 0.2.0",
	}
	got, err := FoundCodex(context.Background(), NewestCodexOpts{
		Home:   home,
		Getenv: func(string) string { return "" },
		IsExecutable: func(p string) bool {
			_, ok := versions[filepath.Clean(p)]
			return ok
		},
		RunVersion: func(ctx context.Context, bin string) (string, error) {
			_ = ctx
			v, ok := versions[filepath.Clean(bin)]
			if !ok {
				return "", fmt.Errorf("no version for %s", bin)
			}
			return v, nil
		},
		RunLogin: loginPATH(loginDir),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("FoundCodex = %#v, want 2 rows", got)
	}
	if got[0].Path != filepath.Clean(local) || got[0].Via != viaWellKnown || got[0].Version != "0.1.0" {
		t.Fatalf("row0 = %#v", got[0])
	}
	if got[1].Path != filepath.Clean(loginBin) || got[1].Via != viaLoginShell || got[1].Version != "0.2.0" {
		t.Fatalf("row1 = %#v", got[1])
	}
}

func TestFoundCodexDedupeWellKnownAndLogin(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	local := filepath.Join(home, ".local", "bin", "codex")
	got, err := FoundCodex(context.Background(), NewestCodexOpts{
		Home:   home,
		Getenv: func(string) string { return "" },
		IsExecutable: func(p string) bool {
			return filepath.Clean(p) == filepath.Clean(local)
		},
		RunVersion: func(ctx context.Context, bin string) (string, error) {
			_ = ctx
			return "codex-cli 0.3.0", nil
		},
		RunLogin: loginPATH(filepath.Join(home, ".local", "bin")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("FoundCodex = %#v, want 1 row", got)
	}
	if got[0].Via != viaWellKnown {
		t.Fatalf("Via = %q, want well_known", got[0].Via)
	}
}

func TestFoundCodexDropsUnversionedAndMissing(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	good := filepath.Join(home, ".local", "bin", "codex")
	got, err := FoundCodex(context.Background(), NewestCodexOpts{
		Home:   home,
		Getenv: func(string) string { return "" },
		IsExecutable: func(p string) bool {
			return filepath.Clean(p) == filepath.Clean(good)
		},
		RunVersion: func(ctx context.Context, bin string) (string, error) {
			_ = ctx
			return "garbage", nil
		},
		RunLogin: loginPATH(""),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("FoundCodex = %#v, want empty", got)
	}
}

func TestFoundCodexNothingUsable(t *testing.T) {
	t.Parallel()
	got, err := FoundCodex(context.Background(), NewestCodexOpts{
		Home:         t.TempDir(),
		Getenv:       func(string) string { return "" },
		IsExecutable: func(string) bool { return false },
		RunVersion:   func(context.Context, string) (string, error) { return "", fmt.Errorf("nope") },
		RunLogin:     loginPATH(""),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %#v", got)
	}
}

func TestNewestCodexPicksHigherVersion(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	older := filepath.Join(home, ".local", "bin", "codex")
	loginDir := filepath.Join(home, "login-bin")
	newer := filepath.Join(loginDir, "codex")
	versions := map[string]string{
		filepath.Clean(older): "0.1.0",
		filepath.Clean(newer): "0.2.0",
	}
	path, ver, err := NewestCodex(context.Background(), NewestCodexOpts{
		Home:   home,
		Getenv: func(string) string { return "" },
		IsExecutable: func(p string) bool {
			_, ok := versions[filepath.Clean(p)]
			return ok
		},
		RunVersion: func(ctx context.Context, bin string) (string, error) {
			_ = ctx
			return versions[filepath.Clean(bin)], nil
		},
		RunLogin: loginPATH(loginDir),
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Clean(newer) || ver != "0.2.0" {
		t.Fatalf("NewestCodex = %q %q, want %q 0.2.0", path, ver, newer)
	}
}

func TestNewestCodexTieKeepsListOrder(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	first := filepath.Join(home, ".local", "bin", "codex")
	loginDir := filepath.Join(home, "login-bin")
	second := filepath.Join(loginDir, "codex")
	path, ver, err := NewestCodex(context.Background(), NewestCodexOpts{
		Home:   home,
		Getenv: func(string) string { return "" },
		IsExecutable: func(p string) bool {
			c := filepath.Clean(p)
			return c == filepath.Clean(first) || c == filepath.Clean(second)
		},
		RunVersion: func(context.Context, string) (string, error) {
			return "0.9.0", nil
		},
		RunLogin: loginPATH(loginDir),
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Clean(first) || ver != "0.9.0" {
		t.Fatalf("tie winner = %q %q, want first well-known %q", path, ver, first)
	}
}

func TestNewestCodexNoneFound(t *testing.T) {
	t.Parallel()
	_, _, err := NewestCodex(context.Background(), NewestCodexOpts{
		Home:         t.TempDir(),
		Getenv:       func(string) string { return "" },
		IsExecutable: func(string) bool { return false },
		RunLogin:     loginPATH(""),
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFoundCodexEnvIsCandidate(t *testing.T) {
	t.Parallel()
	home := t.TempDir()
	envBin := filepath.Join(home, "custom", "codex")
	well := filepath.Join(home, ".local", "bin", "codex")
	versions := map[string]string{
		filepath.Clean(envBin): "0.1.0",
		filepath.Clean(well):   "0.4.0",
	}
	path, ver, err := NewestCodex(context.Background(), NewestCodexOpts{
		Home: home,
		Getenv: func(k string) string {
			if k == EnvCodexBin {
				return envBin
			}
			return ""
		},
		IsExecutable: func(p string) bool {
			_, ok := versions[filepath.Clean(p)]
			return ok
		},
		RunVersion: func(ctx context.Context, bin string) (string, error) {
			_ = ctx
			return versions[filepath.Clean(bin)], nil
		},
		RunLogin: loginPATH(""),
	})
	if err != nil {
		t.Fatal(err)
	}
	if path != filepath.Clean(well) || ver != "0.4.0" {
		t.Fatalf("got %q %q, want well-known newer", path, ver)
	}
}

func writeVersionScript(t *testing.T, dir, stdout string) string {
	t.Helper()
	bin := filepath.Join(dir, "codex")
	body := "#!/bin/sh\nprintf '%s\\n' '" + strings.ReplaceAll(stdout, "'", "") + "'\n"
	if err := os.WriteFile(bin, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func loginPATH(dir string) func(shell, command string, env []string) (string, error) {
	return func(shell, command string, env []string) (string, error) {
		_ = shell
		_ = command
		_ = env
		if dir == "" {
			return "OTHER=1\x00", nil
		}
		return "PATH=" + dir + "\x00", nil
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
