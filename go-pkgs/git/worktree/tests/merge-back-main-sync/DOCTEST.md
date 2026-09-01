# Merge-back: sync main onto upstream before land

## Version
0.0.1

Before landing a linked worktree into main, `worktree.MergeBack` refreshes
**main** from its upstream (`git fetch` + `git rebase` onto the upstream tip).

Upstream resolution:

1. Configured `branch.<name>.remote` + `branch.<name>.merge` when set
2. Else `origin/<current-branch>` when the `origin` remote exists
3. Else **skip** remote sync (local fixtures / no remote)

When a remote is resolved: main must be clean; fetch/rebase failures are fatal
(no land). Dry-run lists fetch/rebase commands without mutating.

# DSN (Domain Specific Notion)

**Participants**

- **Main** — named-branch checkout that receives the land
- **Upstream** — remote-tracking tip main rebases onto (`@{u}` or `origin/<branch>`)
- **Worktree** — linked branch being merged back
- **Remote sync** — pre-land `git fetch` + `git rebase` on main only

**Behaviors**

- When no remote can be resolved, remote sync is skipped and land proceeds as before.
- When a remote is resolved, main must be clean; fetch/rebase failures abort before land.
- Dry-run lists the sync commands without mutating main or the feature worktree.

## Decision Tree

```
merge-back-main-sync
├── no-remote/                    no origin → sync skipped
│   └── ahead-ok/                 ahead land still succeeds
└── with-remote/                  origin present
    ├── behind-success/           main behind origin → sync then rebased-and-merged
    ├── dirty-main-errors/        dirty main → error, no land
    ├── missing-remote-branch-errors/  origin exists but branch missing → error
    └── dry-run-lists-sync/       DryRun prints fetch + rebase
```

## Test Index

| Leaf | Outcome |
|------|---------|
| `no-remote/ahead-ok` | No origin: MergeBack lands ahead branch (Action=merged) |
| `with-remote/behind-success` | Main behind origin: sync then land as `rebased-and-merged`; main has remote-only + feature |
| `with-remote/dirty-main-errors` | Dirty main → error containing `main-sync` |
| `with-remote/missing-remote-branch-errors` | Fetch of missing branch → error |
| `with-remote/dry-run-lists-sync` | DryRun stdout contains `fetch` and `rebase origin/` |

## How to Run

```sh
doctest vet ./git/worktree/tests/merge-back-main-sync
doctest test ./git/worktree/tests/merge-back-main-sync
```

```go
import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

type Request struct {
	WorkRoot   string
	MainRepo   string
	SourcePath string
	DryRun     bool
	Stdout     *bytes.Buffer
}

type Response struct {
	Action string
	Err    string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	opts := worktree.MergeBackOptions{
		SourcePath: req.SourcePath,
		DryRun:     req.DryRun,
		TmpDir:     filepath.Join(req.WorkRoot, ".wrk", "worktrees"),
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return true, nil
		},
	}
	if req.Stdout != nil {
		opts.Stdout = req.Stdout
	}
	result, err := worktree.MergeBack(opts)
	if err != nil {
		return &Response{Err: err.Error()}, nil
	}
	if result == nil {
		return &Response{}, nil
	}
	return &Response{Action: result.Action}, nil
}
```
