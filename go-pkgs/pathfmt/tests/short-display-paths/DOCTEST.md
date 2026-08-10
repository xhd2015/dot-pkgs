# Short Display Paths — `pathfmt.Short` / `pathfmt.ShortFrom` / `pathfmt.Expand`

## Version
0.0.3

Unit-style doc tests for the display-only path formatter in
`github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt`. `Short` / `ShortFrom` shorten
filesystem paths for human-readable output; `Expand` converts `~`-prefixed
display paths back to absolute paths for filesystem operations.

## DSN (Domain Specific Notion)

### Participants

- **`Short`** — pure formatter: accepts a filesystem path string, returns a shorter
  string for human-readable output only. Equivalent to `ShortFrom(path, "")`.
- **`ShortFrom`** — same rules as `Short`, but relative base is an explicit
  `baseDir` (empty → process cwd). Used when the reader process has a different
  working directory (e.g. agent workspace vs serve process cwd).
- **`Expand`** — inverse formatter: accepts a display path (often `~`-prefixed),
  returns an absolute path for filesystem use.
- **`cwd` / base** — process working directory from `os.Getwd()` when baseDir is
  empty, or the explicit baseDir; absolutized and symlink-evaluated before comparison.
- **`evalPath`** — canonicalizes paths via `filepath.EvalSymlinks`; when the leaf
  is missing, walks up to the longest existing prefix and reattaches the suffix
  so cwd-relative shortening still works (e.g. pre-install integration targets,
  macOS `/var` vs `/private/var` cwd mismatch).
- **`home`** — user home directory from `os.UserHomeDir()`, used as a `~` prefix
  when the path is outside the base subtree but still under home. When **base
  equals home**, relative shortening is skipped so under-home paths prefer
  `"~/..."` over `".spl/..."` (agent-safe).

### Behaviors

**Short / ShortFrom**

- **Normalize** — `filepath.Abs(path)`; empty or Abs error → return input unchanged.
- **Base** — `ShortFrom(path, baseDir)`: non-empty baseDir, else `Getwd()`.
- **Under base (base ≠ home)** — path equals base → `"."`; strict child → `rel`
  without `./` prefix (rel must not start with `".."`). Applies to existing files
  **and** missing leaf paths whose existing-prefix canonicalization still lies
  under base.
- **Base is home** — skip relative form; fall through to home shorten.
- **Home shorten** — when not under base (or base is home), if path has prefix
  `home + sep`, replace with `"~" + suffix`.
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
├── short-from                   [ShortFrom: explicit baseDir]
│   ├── base-workspace-under-home   base=workspace, path under ~/.spl → "~/..."
│   ├── base-is-home-skips-rel      base=home, path under home → "~/..." (not .spl/...)
│   ├── base-child-rel              base=temp proj, path=child → "child"
│   └── empty-base-uses-cwd         base="", cwd=proj → same as Short
├── edge-inputs                  [Short: normalization edge cases]
│   ├── empty-path               "" → "" (unchanged)
│   └── outside-home             temp dir outside home → absolute unchanged
├── missing-targets              [Short: missing leaf paths under cwd]
│   ├── missing-file-parent-exists  parent dirs exist, file missing → rel
│   ├── missing-nested-path         fully missing nested path → rel
│   └── cwd-var-prefix-mismatch     /var vs /private/var cwd mismatch → rel
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
| `short-from/base-workspace-under-home` | ShortFrom | base=workspace under home; path `~/.spl/...` → `~/...` |
| `short-from/base-is-home-skips-rel` | ShortFrom | base=home; path under home → `~/...` not `.spl/...` |
| `short-from/base-child-rel` | ShortFrom | base=temp proj; child path → `"child"` |
| `short-from/empty-base-uses-cwd` | ShortFrom | empty baseDir uses process cwd like Short |
| `edge-inputs/empty-path` | Short | Empty input returned unchanged |
| `edge-inputs/outside-home` | Short | Path outside home stays absolute |
| `missing-targets/missing-file-parent-exists` | Short | Missing leaf with parent dirs → cwd-relative |
| `missing-targets/missing-nested-path` | Short | Fully missing nested path under cwd → cwd-relative |
| `missing-targets/cwd-var-prefix-mismatch` | Short | `/var` vs `/private/var` cwd mismatch still shortens |
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

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/pathfmt"
)

type Request struct {
	Path    string
	Op      string // "short" (default), "short-from", or "expand"
	BaseDir string // for Op "short-from"; empty → process cwd
}

type Response struct {
	Display string
	Cwd     string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	var display string
	switch req.Op {
	case "expand":
		display = pathfmt.Expand(req.Path)
	case "short-from":
		display = pathfmt.ShortFrom(req.Path, req.BaseDir)
	default:
		display = pathfmt.Short(req.Path)
	}
	return &Response{Display: display, Cwd: cwd}, nil
}
```
