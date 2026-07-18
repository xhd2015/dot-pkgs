# scan_repo — Scan path: seed, warm-serve, sibling, liveness via repo index (P2)

## Version
0.0.2

Nested doc tests for **Phase 2**: when `CacheRoot` is set and `NoCache=false`,
`Scan` seeds a durable per-universe `repos.json` (universe `"home"` for single-root
library fixtures), warm-serves O(repos) from that index (plus liveness), discovers
**sibling** checkouts under the same parent via `ReadDir` without a full cold
re-walk, and omits indexed paths whose `.git` is gone (liveness on the Scan path).

Depends on P1 (`LoadRepoIndex` / `SaveRepoIndex` / `ApplyLiveness`). Scan wires
index seed/serve/sibling alongside mirror warm (implemented; leaves expect green).

**Out of scope:** walk JSONL, gen_end, adaptive budget tiers, wrk CLI, pure
Load/Save unit trees (see sibling nested `cache/repo-index/`).

Nested `DOCTEST.md` isolates `Request`/`Response`/`Run` from the parent mirror
`CacheOp` contract and from P1 pure index I/O.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — supplies one scan root, explicit temp `CacheRoot`, `NoCache=false`.
- **Scan (cold)** — full live walk; discovers mains; **seeds** universe index at
  `<CacheRoot>/home/repos.json` with `Base` = abs of the scan root (simplest P2
  library contract) and `repos[]` including at least main paths found.
- **Scan (warm)** — when root is warm-eligible (complete root mirror from cold)
  **or** when a usable index exists: serve candidates from the **repo index**
  (not only mirror WalkDir of is_repo marks), apply liveness, merge **sibling**
  discoveries.
- **Repo index store** — schema v1 `repos.json` under universe `home` (P1 helpers).
- **Liveness (Scan path)** — drop / omit entries whose `path/.git` is missing
  (dir or gitfile); dead repos never appear in `Result.Repos`.
- **Sibling probe** — for an indexed (or served) repo at `parent/A`, `ReadDir(parent)`
  finds `parent/B` with `.git` even if B was never cold-written to index/mirror;
  warm Scan includes B without `Refresh`/full re-cold.
- **Filesystem** — fake `.git` fixtures; no enrichment, no git CLI.

### Behaviors

- **Cold seeds index:** first Scan with `CacheRoot` + `!NoCache` writes
  `home/repos.json` containing discovered main paths (at least).
- **Warm serves index:** second Scan returns the indexed live repos (same as cold
  seed set when FS unchanged); index still loadable after warm.
- **Sibling discovers new:** after cold has only `parent/A`, plant `parent/B/.git`;
  warm Scan includes **both** A and B (sibling ReadDir), without `Refresh=true`.
- **Liveness drops dead via Scan:** after cold indexes live + doomed, remove
  `doomed/.git` (or the repo); next Scan omits doomed and keeps live.

## Decision Tree

```
index-serve                    [nested — Scan + home/repos.json + sibling]
├── cold-seed/                 [first Scan writes index]
│   └── writes-index/          # home/repos.json lists cold-discovered mains
├── warm-serve/                [second Scan serves from index]
│   └── from-index/            # Result matches indexed live repos; IndexOK
├── sibling/                   [warm discovers uncached sibling]
│   └── discovers-new/         # plant B after cold A; warm finds A+B
└── liveness/                  [Scan path liveness, not pure ApplyLiveness]
    └── drop-dead-via-scan/    # remove .git of indexed repo → omitted on Scan
```

## Test Index

| Leaf | Phase | Description |
|------|-------|-------------|
| `cold-seed/writes-index` | cold | Cold Scan seeds `home/repos.json` with main paths |
| `warm-serve/from-index` | warm | Warm Scan returns indexed live mains; index still present |
| `sibling/discovers-new` | warm+sibling | Uncached sibling of indexed repo appears without Refresh |
| `liveness/drop-dead-via-scan` | warm+liveness | Dead indexed path omitted from Scan Result |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/cache/index-serve/
doctest test -v ./go-pkgs/git/scan_repo/tests/cache/index-serve/
```

From monorepo / worktree with nested external:

```sh
doctest vet ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/index-serve/
doctest test -v ./external/dot-pkgs-master-2026-07-15/go-pkgs/git/scan_repo/tests/cache/index-serve/
```

```go
import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

// P2 library universe for single-root fixtures: always "home",
// Base = abs of first root (see requirement simplest P2).
const indexUniverseHome = "home"

type Request struct {
	Roots     []string
	CacheRoot string
	NoCache   bool
	Refresh   bool
	Debug     bool
	// DeadPath is the abs path of a repo removed or stripped of .git (liveness leaf).
	DeadPath string
	// KnownPath / SiblingPath stash abs paths for Assert (optional).
	KnownPath   string
	SiblingPath string
}

type Response struct {
	Repos      []scan_repo.Repo
	RootErrors []scan_repo.RootError
	// Index is LoadRepoIndex(CacheRoot, "home") after Scan.
	Index     scan_repo.RepoIndex
	IndexOK   bool
	IndexPath string
	DebugOut  string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	var stderr bytes.Buffer
	opts := scan_repo.Options{
		Roots:     req.Roots,
		CacheRoot: req.CacheRoot,
		NoCache:   req.NoCache,
		Refresh:   req.Refresh,
		Debug:     req.Debug,
	}
	if req.Debug {
		opts.Stderr = &stderr
	}
	result, err := scan_repo.Scan(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	idx, ok, loadErr := scan_repo.LoadRepoIndex(req.CacheRoot, indexUniverseHome)
	if loadErr != nil {
		return nil, loadErr
	}
	return &Response{
		Repos:      result.Repos,
		RootErrors: result.RootErrors,
		Index:      idx,
		IndexOK:    ok,
		IndexPath:  filepath.Join(req.CacheRoot, indexUniverseHome, "repos.json"),
		DebugOut:   stderr.String(),
	}, nil
}
```
