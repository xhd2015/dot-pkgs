# Short Display Paths — `pathfmt.Short` / `pathfmt.Expand`

## Version
0.0.2

Unit-style doc tests for the display-only path formatter in
`github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt`. `Short` shortens filesystem paths
for human-readable CLI output; `Expand` converts `~`-prefixed display paths back to
absolute paths for filesystem operations.

## DSN (Domain Specific Notion)

### Participants

- **`Short`** — pure formatter: accepts a filesystem path string, returns a shorter
  string for human-readable output only.
- **`Expand`** — inverse formatter: accepts a display path (often `~`-prefixed),
  returns an absolute path for filesystem use.
- **`cwd`** — process working directory from `os.Getwd()`, absolutized and
  symlink-evaluated before comparison.
- **`home`** — user home directory from `os.UserHomeDir()`, used as a `~` prefix
  when the path is outside the cwd subtree but still under home.

### Behaviors

**Short**

- **Normalize** — `filepath.Abs(path)`; empty or Abs error → return input unchanged.
- **Under cwd** — path equals cwd → `"."`; strict child → `rel` without `./` prefix
  (rel must not start with `".."`).
- **Home shorten** — when not under cwd, if path has prefix `home + sep`, replace
  with `"~" + suffix`.
- **Fallback** — return the absolute path unchanged.

**Expand**

- **Passthrough** — empty or no `~` prefix → return input unchanged.
- **Tilde home** — `"~"` → home directory.
- **Tilde subpath** — `"~/..."` → `filepath.Join(home, suffix)`.
- **Home error** — `UserHomeDir` error → return input unchanged.

## Decision Tree

```
short-display-paths
├── cwd-relative                 [Short: path under cwd subtree]
│   ├── at-cwd                   path == cwd → "."
│   ├── child-path               strict child → "child"
│   ├── deep-nested-child        nested child → "a/b/c"
│   ├── relative-input           relative input under cwd → "child"
│   └── parent-path              parent of cwd → ~ or absolute (not "./..")
├── home-shorten                 [Short: outside cwd, under home]
│   ├── at-home                  path == home → "~"
│   └── under-home               cache-like path → "~/..."
├── edge-inputs                  [Short: normalization edge cases]
│   ├── empty-path               "" → "" (unchanged)
│   └── outside-home             temp dir outside home → absolute unchanged
└── expand-path                  [Expand: ~ display paths → absolute]
    ├── tilde-home               "~" → absolute home
    ├── tilde-subpath            "~/foo/bar" → join(home, "foo", "bar")
    ├── non-tilde                absolute path → unchanged
    └── empty                    "" → "" (unchanged)
```

## Test Index

| Leaf | Op | Description |
|------|-----|-------------|
| `cwd-relative/at-cwd` | Short | Absolute path equal to cwd displays as `"."` |
| `cwd-relative/child-path` | Short | Strict child of cwd displays as `"child"` |
| `cwd-relative/deep-nested-child` | Short | Deeply nested child displays as `"a/b/c"` |
| `cwd-relative/relative-input` | Short | Relative input resolved under cwd displays as `"child"` |
| `cwd-relative/parent-path` | Short | Parent of cwd is not shortened to `./..`; uses `~` or absolute |
| `home-shorten/at-home` | Short | Home directory displays as `"~"` when cwd is elsewhere |
| `home-shorten/under-home` | Short | Path under home (not cwd) displays with `~` prefix |
| `edge-inputs/empty-path` | Short | Empty input returned unchanged |
| `edge-inputs/outside-home` | Short | Path outside home stays absolute |
| `expand-path/tilde-home` | Expand | `"~"` expands to absolute home directory |
| `expand-path/tilde-subpath` | Expand | `"~/foo/bar"` expands to `filepath.Join(home, "foo", "bar")` |
| `expand-path/non-tilde` | Expand | Non-tilde absolute path returned unchanged |
| `expand-path/empty` | Expand | Empty input returned unchanged |

## How to Run

```sh
doctest vet ./external/dot-pkgs/go-pkgs/pathfmt/tests/short-display-paths/
doctest test ./external/dot-pkgs/go-pkgs/pathfmt/tests/short-display-paths/
```

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

type Request struct {
	Path string
	Op   string // "short" (default) or "expand"
}

type Response struct {
	Display string
	Cwd     string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var display string
	switch req.Op {
	case "expand":
		display = pathfmt.Expand(req.Path)
	default:
		display = pathfmt.Short(req.Path)
	}
	return &Response{Display: display, Cwd: cwd}, nil
}
```