# scan_repo — durable per-universe repo index (P1)

## Version
0.0.2

Nested doc tests for the **repo index** store: durable `repos.json` under
`CacheRoot` for universe `home` or `root`, with load/save (atomic write) and
liveness drop of entries whose `.git` is gone.

This is Classic TDD P1 only — pure library I/O. No Scan warm path, no walk
JSONL, no sibling probe, no wrk CLI.

Nested `DOCTEST.md` isolates `Request`/`Response`/`Run` from the parent mirror
`CacheOp` contract (`SaveCacheEntry` / `LoadCacheEntry`).

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies explicit temp `CacheRoot`, universe (`home`|`root`), and
  an in-memory `RepoIndex` (or empty for load-missing).
- **Repo index store** — files at `<CacheRoot>/home/repos.json` and
  `<CacheRoot>/root/repos.json` (or equivalent UniverseHome/UniverseRoot helpers).
- **Schema v1 document** — `version`, `universe`, `base` (abs path for the
  universe root), `updated_at`, and `repos[]` with `path`, `repo_type`,
  `git_dir`, `depth`, `seen_at`.
- **SaveRepoIndex** — persists index under the universe path with atomic rename
  (and flock where the product uses it); creates parent dirs as needed.
- **LoadRepoIndex** — reads universe file; missing file is empty/not-found, not
  an error.
- **ApplyLiveness** — filters index entries: drop any whose path no longer has
  a live `.git` (directory or gitfile); keep entries that still look like repos.
- **Filesystem** — real or fake `.git` markers used only by liveness leaves.

### Behaviors

- Save then Load of the same universe returns the same schema fields (round-trip).
- Load when `repos.json` is absent returns ok=false (or empty index) with nil error.
- ApplyLiveness removes seeded paths that lack `.git` and retains paths that
  still have `.git`.
- Universes are independent: `home` and `root` files do not overwrite each other.

## Decision Tree

```
repo-index                     [nested — Load/SaveRepoIndex + ApplyLiveness]
├── save-load/                 [IndexOp=save-load — round-trip fields]
│   ├── home/                  # universe=home → CacheRoot/home/repos.json
│   └── root/                  # universe=root → CacheRoot/root/repos.json
├── load/                      [IndexOp=load]
│   └── missing/               # no repos.json → empty / ok=false, err nil
└── liveness/                  [IndexOp=liveness]
    └── drop-dead/             # dead path dropped; live fake-git kept
```

## Test Index

| Leaf | IndexOp | Universe | Description |
|------|---------|----------|-------------|
| `save-load/home` | save-load | home | Round-trip all v1 fields under home/ |
| `save-load/root` | save-load | root | Round-trip all v1 fields under root/ |
| `load/missing` | load | home | Missing file → empty index, not error |
| `liveness/drop-dead` | liveness | home | ApplyLiveness drops missing-.git; keeps live |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/cache/repo-index/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/repo-index/
```

From monorepo / worktree with nested external:

```sh
doctest vet ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/repo-index/
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/repo-index/
```

```go
import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

// IndexOp values: "save-load" | "load" | "liveness".
// Empty is invalid for this tree.

type Request struct {
	CacheRoot string
	Universe  string // "home" | "root"
	IndexOp   string // "save-load" | "load" | "liveness"
	// Index is the document to Save (save-load) or to filter (liveness).
	Index scan_repo.RepoIndex
}

type Response struct {
	Index   scan_repo.RepoIndex
	IndexOK bool // true when Load found a file (load / save-load after save)
	// IndexPath is the on-disk repos.json path expected for CacheRoot+Universe.
	IndexPath string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	switch req.IndexOp {
	case "save-load":
		if err := scan_repo.SaveRepoIndex(req.CacheRoot, req.Index); err != nil {
			return nil, err
		}
		idx, ok, err := scan_repo.LoadRepoIndex(req.CacheRoot, req.Universe)
		if err != nil {
			return nil, err
		}
		path := expectedRepoIndexPath(t, req.CacheRoot, req.Universe)
		return &Response{Index: idx, IndexOK: ok, IndexPath: path}, nil
	case "load":
		idx, ok, err := scan_repo.LoadRepoIndex(req.CacheRoot, req.Universe)
		if err != nil {
			return nil, err
		}
		path := expectedRepoIndexPath(t, req.CacheRoot, req.Universe)
		return &Response{Index: idx, IndexOK: ok, IndexPath: path}, nil
	case "liveness":
		// ApplyLiveness filters entries whose .git is gone.
		// Product may take (RepoIndex) or (context, RepoIndex); tests call the
		// unary form. Implementer may adapt the signature to match package style.
		out := scan_repo.ApplyLiveness(req.Index)
		path := expectedRepoIndexPath(t, req.CacheRoot, req.Universe)
		return &Response{Index: out, IndexOK: true, IndexPath: path}, nil
	default:
		return nil, fmt.Errorf("unknown IndexOp %q", req.IndexOp)
	}
}

// expectedRepoIndexPath is the P1 path contract used by Assert/Setup:
// <cacheRoot>/<universe>/repos.json
func expectedRepoIndexPath(t *testing.T, cacheRoot, universe string) string {
	t.Helper()
	return filepath.Join(cacheRoot, universe, "repos.json")
}
```
