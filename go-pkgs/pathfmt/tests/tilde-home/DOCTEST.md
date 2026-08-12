# TildeHome — `pathfmt.TildeHome`

## Version
0.0.2

Unit-style doc tests for display-only home-tilde shortening in
`github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt`. `TildeHome` replaces the user
home directory prefix with `"~"` and never produces cwd-relative forms.

**Classic TDD:** `TildeHome` does not exist yet. Root `Run` calls
`pathfmt.TildeHome` so the suite is compile-RED until the implementer adds the
API. Do not implement production code in this design pass.

## DSN (Domain Specific Notion)

### Participants

- **`TildeHome`** — pure display formatter: accepts a filesystem path string,
  returns a shorter string for human-readable output only. Must not be used for
  file I/O or `exec.Command.Dir`.
- **`path`** — input path (empty, absolute, or relative). Normalized with
  `filepath.Abs` before home rules apply.
- **`home`** — user home from `os.UserHomeDir()`, matched with the same raw
  string-prefix rule as the home step inside `Short` / `ShortFrom` (no extra
  symlink/`evalPath` rules for the home prefix in this API).
- **`cwd`** — process working directory. **Not** used for shortening. Present
  only so leaves can construct paths that sit under both cwd and home, and so
  asserts can prove the result is **not** a cwd-relative form.

### Behaviors

- **Empty** — `""` → `""` unchanged.
- **Normalize** — `filepath.Abs(path)`; on Abs error → return input unchanged.
- **At home** — abs path equals home → `"~"`.
- **Under home** — abs has prefix `home + sep` → `"~" + strings.TrimPrefix(abs, home)`
  (native separators after `~`, same as existing Short home step).
- **Outside home** — return absolute path (no leading `~`).
- **No cwd-relative** — never `"."` or `"child/..."` even when path is a strict
  child of process cwd. Critical distinction from `Short`:
  - `cwd = ~/proj`, `path = ~/proj/skills/foo`
  - `Short` → `"skills/foo"`
  - `TildeHome` → `"~/proj/skills/foo"` (platform-native seps)
- **Relative inputs** — Abs first, then the same home rules.
- **Home unreadable** — `UserHomeDir` error → return abs (or original on Abs
  failure). Not exercised as a leaf (would require env mutation; forbidden).

### Inverse

Existing `Expand` is the inverse of tilde display paths; this tree does not
redesign `Expand`.

## Decision Tree

```
tilde-home                           [TildeHome: home tilde only, no cwd-rel]
├── empty                            "" → ""
├── at-home                          abs home → "~"
├── under-home                       abs under home → "~/..." (exact home step)
├── outside-home                     abs outside home → absolute, no ~
├── under-cwd-and-home               under cwd AND home → "~/..." not rel
└── relative-under-home              relative Abs'd under home → "~/..."
```

### Parameter significance (high → low)

1. **Path relation to home / empty** — empty vs at-home vs under-home vs outside
   (primary display outcome).
2. **Cwd membership** — under both cwd and home proves no cwd-relative form
   (regression vs `Short`).
3. **Input shape** — absolute vs relative (Abs then same rules).

## Test Index

| Leaf | Description |
|------|-------------|
| `empty` | Empty input returned unchanged |
| `at-home` | Absolute home directory displays as `"~"` |
| `under-home` | Absolute path under home → `"~" + TrimPrefix(abs, home)` |
| `outside-home` | Path outside home stays absolute (no `~` prefix) |
| `under-cwd-and-home` | Path under both cwd and home → `~/...`, not cwd-relative |
| `relative-under-home` | Relative input Abs'd under home → `~/...` |

## How to Run

```sh
doctest vet ./external/dot-pkgs-master-2026-08-12-2/go-pkgs/pathfmt/tests/tilde-home
doctest test ./external/dot-pkgs-master-2026-08-12-2/go-pkgs/pathfmt/tests/tilde-home
```

From the go-pkgs module root (package is local):

```sh
doctest vet ./pathfmt/tests/tilde-home
doctest test ./pathfmt/tests/tilde-home
```

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

type Request struct {
	Path string
}

type Response struct {
	Display string
	Cwd     string // diagnostic only; TildeHome must not depend on cwd for correctness
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	display := pathfmt.TildeHome(req.Path)
	return &Response{Display: display, Cwd: cwd}, nil
}
```
