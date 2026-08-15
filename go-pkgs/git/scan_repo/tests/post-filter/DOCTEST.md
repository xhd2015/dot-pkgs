# scan_repo — post-process base-path filter (P1 resolve then filter)

## Version
0.0.2

Nested doc tests for **Plan P1**: after optional worktree resolve
(`ListWorktrees`), `Scan` must post-process so **returned** values never
include repos (or inner worktree paths) outside each scan root.

**Order:** resolve worktrees (when flagged) **then** base-path filter.
Base progress (warm/cold/walk-log consume, index seed/serve) stays as-is;
this tree asserts **return-value** leak-proofing only.

**Classic TDD:** filter-after-resolve is not implemented (or incomplete) in
product `Scan` — new leaves expect **RED** until the implementer lands.

**Out of scope:** wrk `--done` cascade (P2), changing walk-log consume
internals beyond return filter, promoting worktrees to top-level `Repos`.

Nested `DOCTEST.md` isolates `Request`/`Response`/`Run` so walk-log clock
injection and `ListWorktrees` enrichment coexist without redefining the
parent tree contracts.

## DSN (Domain Specific Notion)

### Participants

- **Caller** — one scan root, explicit temp `CacheRoot` (never product home
  cache), optional `ListWorktrees`, optional warm clocks for walk-log budget.
- **Scan (warm + walk-log consume)** — serves index; consumes `walk.jsonl`
  under adaptive budget; may re-list prior visit dirs and discover checkouts
  **outside** the scan root if visits were foreign (core leak).
- **Scan (ListWorktrees)** — `git worktree list --porcelain` on main rows
  fills `Repo.Worktrees`; without filter, outer worktrees leak into the
  return shape.
- **Base-path filter (under test)** — after resolve, every top-level
  `Result.Repos[].Path` and every `Repo.Worktrees[].Path` must lie under
  the abs scan root (`pathIsUnderRoot` / equivalent). Foreign paths stripped
  or never attached.
- **Filesystem** — fake `.git` for walk-log / top-level filter leaves; real
  `git` for ListWorktrees leaves (skip when unavailable).

### Behaviors

1. **Walk-log foreign leak blocked:** warm Scan of consumer only, with
   walk log seeded so consume re-lists a **foreign** parent dir that holds
   `agent-pro`, must **not** return that foreign checkout in
   `Result.Repos` (nor any of its worktrees).
2. **ListWorktrees inner only under root:** main + linked worktree both
   under scan root; `ListWorktrees=true` → linked path appears on
   `main.Worktrees` (inner field). Dual discovery as a separate top-level
   row from FS walk (Option A) is allowed when the path is under root;
   ListWorktrees must not invent extra top-level rows beyond FS discovery.
3. **ListWorktrees outside base stripped:** main under scan root; linked
   worktree **outside** scan root; `ListWorktrees=true` → outer path
   absent from `Worktrees` (only under-root worktrees remain, typically
   the main row itself with `IsMain=true`).
4. **ListWorktrees false still filters top-level:** `ListWorktrees=false`
   → no expand; `Worktrees` empty/nil; under-root filter still applies to
   top-level `Repos` (neighbor outside root never returned).

## Decision Tree

```
post-filter                         [nested — resolve then base-path filter]
├── walk-log-foreign-leak/          # warm consume must not emit foreign agent-pro
├── list-worktrees-inner-only/      # under-root wt on main.Worktrees
├── list-worktrees-outside-base-stripped/  # outer wt stripped from Worktrees
└── list-worktrees-false-top-level-filter/ # flag off: Worktrees empty + top-level filter
```

## Test Index

| Leaf | Mode | Description |
|------|------|-------------|
| `walk-log-foreign-leak` | warm+walk-log | Consume re-list of foreign parent must not return agent-pro |
| `list-worktrees-inner-only` | ListWorktrees | Linked under root listed on main.Worktrees |
| `list-worktrees-outside-base-stripped` | ListWorktrees | Worktree outside scan root stripped from Worktrees |
| `list-worktrees-false-top-level-filter` | filter only | ListWorktrees=false; sibling outside root omitted; Worktrees empty |

## How to Run

```sh
doctest vet ./go-pkgs/git/scan_repo/tests/post-filter/
doctest test -v ./go-pkgs/git/scan_repo/tests/post-filter/
```

From monorepo / worktree with nested external:

```sh
doctest vet ./external/dot-pkgs-master-2026-07-22/go-pkgs/git/scan_repo/tests/post-filter/
doctest test -v ./external/dot-pkgs-master-2026-07-22/go-pkgs/git/scan_repo/tests/post-filter/
```

```go
import (
	"context"
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

// Request drives Scan for P1 post-filter leaves.
// Nested tree owns Request/Response/Run (parent root tree is not redefined).
type Request struct {
	Roots         []string
	CacheRoot     string
	NoCache       bool
	Refresh       bool
	ListRemotes   bool
	ListWorktrees bool

	// WarmRefreshBudget: walk-log leaves use -1 to isolate consume from unit rewalk.
	// 0 → product default on other leaves when cache is warm-eligible.
	WarmRefreshBudget time.Duration

	// LastScanEnd / Now inject walk-consume budget (P4). When SetLastScanEnd,
	// Options.LastScanEnd is set; when SetNow, Options.Now returns NowAt.
	LastScanEnd    time.Time
	SetLastScanEnd bool
	NowAt          time.Time
	SetNow         bool

	// Stashed abs paths for Assert (filled by leaf Setup).
	ConsumerPath string // scan root / under-root checkout
	ForeignPath  string // outside-root path that must never leak
	MainPath     string // main checkout abs
	WorktreePath string // linked worktree abs (under or outside root)
	SiblingPath  string // neighbor checkout outside scan root (flag-off leaf)
}

type Response struct {
	Repos      []scan_repo.Repo
	RootErrors []scan_repo.RootError
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	opts := scan_repo.Options{
		Roots:             req.Roots,
		CacheRoot:         req.CacheRoot,
		NoCache:           req.NoCache,
		Refresh:           req.Refresh,
		ListRemotes:       req.ListRemotes,
		ListWorktrees:     req.ListWorktrees,
		WarmRefreshBudget: req.WarmRefreshBudget,
	}
	if req.SetLastScanEnd {
		opts.LastScanEnd = req.LastScanEnd
	}
	if req.SetNow {
		nowAt := req.NowAt
		opts.Now = func() time.Time { return nowAt }
	}
	result, err := scan_repo.Scan(context.Background(), opts)
	if err != nil {
		return nil, err
	}
	return &Response{Repos: result.Repos, RootErrors: result.RootErrors}, nil
}
```
