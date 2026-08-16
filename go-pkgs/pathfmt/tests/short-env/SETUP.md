# Scenario

**Feature**: `pathfmt.ShortEnv` / `ShortEnvFrom` shorten paths using env path aliases

```
# ShortEnvFrom pipeline
caller path + env[] -> ShortEnvFrom
  -> Abs normalize
  -> build eligible aliases from env
  -> longest segment-boundary prefix -> $NAME + rest
  -> else TildeHome(path)

# ShortEnv wrapper
caller path -> ShortEnv -> ShortEnvFrom(path, os.Environ())
```

## Preconditions

- The `pathfmt` package is importable (`github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt`).
- `ShortEnv` / `ShortEnvFrom` are display-only; do not use results for I/O.
- Parallel-safe: no `os.Chdir` / `t.Chdir` / `os.Setenv` / `t.Setenv` in harness.
  Env is injected via `req.Env` for `Op=from`. Process cwd is undetermined.
- Leaves read `os.UserHomeDir()` **read-only** for home cases. Synthetic
  prefixes use `t.TempDir()` (absolute, typically outside home).
- Suite is expected compile-RED until `ShortEnv` / `ShortEnvFrom` exist.

## Steps

1. Leaf / grouping `Setup` sets `req.Op`, `req.Path`, and (for `from`) `req.Env`.
2. Root `Run` calls `pathfmt.ShortEnvFrom` or `pathfmt.ShortEnv` per `req.Op`.

## Context

- Platform-native `filepath` separators appear after `$NAME` and after `~`.
- Shared helpers resolve home/abs and build `KEY=value` pairs without mutating
  process state.
- Fallback expectations use the same home-step rules as `TildeHome` /
  `pathfmt.TildeHome` (string prefix on Abs path).

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

// expectedTildeHome mirrors pathfmt.TildeHome home-step display form.
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

func envPair(key, value string) string {
	return key + "=" + value
}

// mustCwdUnderHome returns Abs(cwd) and Abs(home) when cwd lies under home.
// Skips otherwise — dual membership cannot be forced without forbidden chdir.
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

// expectedEnvDisplay builds $NAME + remainder when abs has prefix at a
// segment boundary. Used by replace leaves for exact expected strings.
func expectedEnvDisplay(name, prefix, abs string) string {
	if abs == prefix {
		return "$" + name
	}
	sep := string(filepath.Separator)
	if strings.HasPrefix(abs, prefix+sep) {
		return "$" + name + strings.TrimPrefix(abs, prefix)
	}
	return abs
}

// assertNoDollarVar fails if display starts with $name as an env-style prefix.
func assertNoDollarVar(t *testing.T, display, name string) {
	t.Helper()
	if display == "$"+name || strings.HasPrefix(display, "$"+name+"/") || strings.HasPrefix(display, "$"+name+`\`) {
		t.Fatalf("display must not use $%s: got %q", name, display)
	}
}
```
