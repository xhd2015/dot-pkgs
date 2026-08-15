# git/reposnapshot — Repo Tree Snapshot Builder

## Version
0.0.2

Doc tests for `github.com/xhd2015/dot-pkgs/go-pkgs/git/reposnapshot`. `Build`
converts a `scan_repo.Result` into nested `Node` trees (main + linked worktrees)
with checkout enrichment. Root scan failures become synthetic error nodes.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies `scan_repo.Result` and `rel(absPath)` callback for paths.
- **Builder** — groups main repos, nests linked worktrees under main nodes, enriches
  each checkout via `checkout.Enrich`.
- **Scan result** — `Repos` rows from discovery; `RootErrors` for failed roots.
- **Node** — `Path`, `Checkout` meta, nested `Worktrees`, merged `Error`.

### Behaviors

- **Main + worktree** — one `Node` per main repo; linked worktrees nested in
  `Worktrees` sorted by path; worktree-only scan rows do not become top-level nodes.
- **Root error** — `RootErrors` recorded in `Snapshot.RootErrors` and as synthetic
  `Node` with `Error` prefixed `scan failed: `.
- Paths use caller's `rel` callback (e.g. home-relative for backup).

## Decision Tree

```
reposnapshot
└── build/
    ├── main-and-worktree/
    └── root-error/
```

## Test Index

| Leaf | Description |
|------|-------------|
| `build/main-and-worktree` | Scan main+linked → nested nodes with checkout meta |
| `build/root-error` | RootErrors → synthetic node + RootErrors entry |

## How to Run

```sh
doctest vet ./go-pkgs/git/reposnapshot/tests/
doctest test -v ./go-pkgs/git/reposnapshot/tests/
```

```go
import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/reposnapshot"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

type Request struct {
	Mode           string // "scan" | "manual"
	ScanRoots      []string
	BaseDir        string
	ManualResult   scan_repo.Result
}

type Response struct {
	Snapshot reposnapshot.Snapshot
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	rel := func(abs string) string {
		rel, err := filepath.Rel(req.BaseDir, abs)
		if err != nil {
			return filepath.ToSlash(abs)
		}
		return filepath.ToSlash(rel)
	}

	var result scan_repo.Result
	switch req.Mode {
	case "scan":
		var err error
		result, err = scan_repo.Scan(context.Background(), scan_repo.Options{
			Roots:         req.ScanRoots,
			ListWorktrees: true,
		})
		if err != nil {
			return nil, err
		}
	case "manual":
		result = req.ManualResult
	default:
		t.Fatalf("unknown mode %q", req.Mode)
	}

	snap := reposnapshot.Build(result, rel)
	return &Response{Snapshot: snap}, nil
}
```