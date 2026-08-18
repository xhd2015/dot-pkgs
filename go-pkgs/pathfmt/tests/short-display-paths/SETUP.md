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
- Leaves set `req.BaseDir` and use `Op=short-from`. No process `chdir`.

## Steps

1. Ancestor `Setup` chains configure `req.Path`, `req.Op`, and `req.BaseDir`.
2. Root `Run` calls `pathfmt.Short`, `pathfmt.ShortFrom`, or `pathfmt.Expand` per `req.Op`.

## Context

- Platform-native `filepath` separators are used in expectations.
- Grouping nodes under `cwd-relative` and `home-shorten` set `req.BaseDir`
  on a temp dir; leaves set concrete paths.
- `expand-path` leaves set `req.Op` to `"expand"`.

```go
import (
	"os"
	"testing"
)

func mkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
}
```