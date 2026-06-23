# Scenario

**Feature**: `pathfmt.Short` and `pathfmt.Expand` format filesystem paths for display and I/O

```
# Short pipeline
caller path string -> Short -> Abs normalize -> cwd/home rules -> display string

# Expand pipeline
display path string -> Expand -> ~ rules -> absolute path

# cwd rules (Short, checked first)
path == cwd -> "." | strict child of cwd -> rel (no ./ prefix)

# home shorten (Short, when not under cwd)
path under home -> "~" + suffix | otherwise -> absolute unchanged
```

## Preconditions

- The `pathfmt` package is importable (`github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt`).
- `Short` is display-only; `Expand` is for converting display paths back to absolute.
- Ancestor `Setup` functions may `chdir` to control cwd; root saves and restores
  the original working directory.

## Steps

1. Ancestor `Setup` chains configure `req.Path`, `req.Op`, and optionally change cwd.
2. Root `Run` calls `pathfmt.Short` or `pathfmt.Expand` per `req.Op` and records cwd.

## Context

- Platform-native `filepath` separators are used in expectations.
- Grouping nodes under `cwd-relative` and `home-shorten` create temp dirs and
  chdir before leaves set concrete paths.
- `expand-path` leaves set `req.Op` to `"expand"` and do not require cwd manipulation.

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func saveAndRestoreCwd(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})
	return wd
}

func chdirTo(t *testing.T, dir string) {
	t.Helper()
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(abs); err != nil {
		t.Fatal(err)
	}
}

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}
```