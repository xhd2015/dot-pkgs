# Scenario

**Feature**: `pathfmt.TildeHome` shortens paths with a home `~` prefix only

```
# TildeHome pipeline (no cwd-relative branch)
caller path string -> TildeHome -> Abs normalize -> home rules -> display string

# home rules
empty -> "" | abs == home -> "~" | under home -> "~" + suffix | else -> abs
```

## Preconditions

- The `pathfmt` package is importable (`github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt`).
- `TildeHome` is display-only; do not use results for I/O or `exec.Command.Dir`.
- Parallel-safe: no `os.Chdir` / `t.Chdir` / `os.Setenv` / `t.Setenv` in harness
  or product isolation. Leaves read `os.UserHomeDir()` and optionally
  `os.Getwd()` only to build inputs and check **non**-relative form.
- Process cwd is undetermined. Expected values for home rules use
  `filepath.Abs` + `UserHomeDir` string prefix (same as Short's home step).
  Leaves that need a path under both cwd and home skip when cwd is not under
  home (cannot construct dual membership without forbidden chdir).

## Steps

1. Leaf `Setup` sets `req.Path` (empty, home, under-home, outside-home, or
   relative).
2. Root `Run` calls `pathfmt.TildeHome(req.Path)` and records cwd for diagnostics.

## Context

- Platform-native `filepath` separators appear after `~` (same as Short home step:
  `"~" + strings.TrimPrefix(abs, home)`).
- Shared helpers resolve home/cwd without mutating process state.

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustUserHome(t *testing.T) string {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(home) == "" {
		t.Fatal("UserHomeDir returned empty home")
	}
	return home
}

func absPath(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func isUnderHome(abs, home string) bool {
	if abs == home {
		return true
	}
	return strings.HasPrefix(abs, home+string(filepath.Separator))
}

// mustCwdUnderHome returns Abs(cwd) and Abs(home) when cwd lies under home.
// Skips the leaf otherwise — dual membership cannot be forced without chdir.
func mustCwdUnderHome(t *testing.T) (cwdAbs, homeAbs string) {
	t.Helper()
	homeAbs = absPath(t, mustUserHome(t))
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	cwdAbs = absPath(t, cwd)
	if !isUnderHome(cwdAbs, homeAbs) {
		t.Skipf("cwd %q not under home %q; skip dual-membership leaf without chdir", cwdAbs, homeAbs)
	}
	return cwdAbs, homeAbs
}

// mustCwdStrictChildOfHome is like mustCwdUnderHome but requires cwd != home so
// Short would prefer a cwd-relative form; TildeHome must still return ~/....
func mustCwdStrictChildOfHome(t *testing.T) (cwdAbs, homeAbs string) {
	t.Helper()
	cwdAbs, homeAbs = mustCwdUnderHome(t)
	if cwdAbs == homeAbs {
		t.Skipf("cwd equals home %q; Short would also use ~ form — skip cwd-rel regression", homeAbs)
	}
	return cwdAbs, homeAbs
}

// expectedTildeHome is the home-step display form matching Short's home branch:
// abs == home -> "~"; under home -> "~" + TrimPrefix(abs, home); else abs.
func expectedTildeHome(abs, home string) string {
	if abs == home {
		return "~"
	}
	homePrefix := home + string(filepath.Separator)
	if strings.HasPrefix(abs, homePrefix) {
		return "~" + strings.TrimPrefix(abs, home)
	}
	return abs
}
```
