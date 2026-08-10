# MergeBack inject options — TmpDir / StashLabel (P4)

Doc-style tests proving that `worktree.MergeBackOptions` accepts injectable
**tmp worktree parent directory** and **stash message label**, and that library
defaults are **product-neutral** (no hard-coded `WRK_HOME` / `"wrk-merge-back"`
required for the dirty-diverged tmp-worktree path).

Callers such as wrk may still pass wrk-specific values; the library must not
require them.

Import path under test:

`github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree`

# DSN (Domain Specific Notion)

**Participants**

- **Caller** — harness (or wrk) building `MergeBackOptions` with optional
  inject fields.
- **MergeBack** — orchestrates dirty-diverged rebase via a temporary linked
  worktree, then migrates uncommitted changes with stash.
- **MergeBackOptions** — public config: `SourcePath`, `TargetPath`, `Remove`,
  `Confirm`, plus injectables **`TmpDir`** (parent dir for tmp worktrees) and
  **`StashLabel`** (message for `git stash push -m`).
- **Tmp worktree** — throwaway linked checkout created under `TmpDir` (or a
  neutral default such as `os.TempDir()` when unset).
- **Stash** — repo-wide stash entries used to probe dirty-change conflicts
  after rebase; labeled with `StashLabel` (or a neutral default when unset).

**Behaviors**

- When `TmpDir` is set, the tmp worktree path used for rebase is a child of
  that directory (visible on `MergeBackPlan.Commands` as the rebase `-C` dir
  before confirm proceeds).
- When `StashLabel` is set, dirty-diverged migration uses that label for
  `stash push -m` (and for recognizing the created entry); stash reflog retains
  the label after pop/drop.
- When both inject fields are empty, dirty-diverged still succeeds **without**
  reading `WRK_HOME` and **without** requiring the product string
  `"wrk-merge-back"`: tmp parent is neutral (e.g. under `os.TempDir()`); stash
  label is not product-specific.

## Version

0.0.2

## Decision Tree

```
merge-back-inject
├── tmp-dir/
│   └── custom-parent/          (LEAF) inject TmpDir → rebase plan dir under it
├── stash-label/
│   └── custom-message/         (LEAF) inject StashLabel → stash reflog has label
└── defaults/
    └── neutral/                (LEAF) empty injects → no WRK_HOME / no wrk-merge-back
```

## Test Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | `tmp-dir/custom-parent` | Dirty diverged + custom `TmpDir`: Confirm sees rebase cmd dir under inject parent; merge succeeds; tmp cleaned |
| 2 | `stash-label/custom-message` | Dirty diverged + custom `StashLabel`: succeeds; stash reflog contains the label |
| 3 | `defaults/neutral` | Empty `TmpDir`/`StashLabel`, no env: succeeds; observed tmp under `os.TempDir()`; stash label not `"wrk-merge-back"` |

## How to Run

```sh
cd external/dot-pkgs-master-2026-07-30/go-pkgs
doctest vet ./git/worktree/tests/merge-back-inject
doctest test ./git/worktree/tests/merge-back-inject/...
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

// Request drives MergeBack with optional inject fields.
// All leaves use diverged + dirty + Remove=false (tmp-worktree path).
type Request struct {
	WorkRoot   string
	MainRepo   string
	SourcePath string
	TargetPath string // empty → main repo checkout

	// TmpDir is passed to MergeBackOptions.TmpDir (tmp worktree parent).
	// Empty means library default (must be product-neutral).
	TmpDir string

	// StashLabel is passed to MergeBackOptions.StashLabel.
	// Empty means library default (must not require "wrk-merge-back").
	StashLabel string
}

type Response struct {
	Action   string
	Branch   string
	Relation string

	// ObservedTmpPath is the rebase worktree path from plan.Commands[0].Dir
	// captured in Confirm (dirty-diverged tmp path builds the plan before confirm).
	ObservedTmpPath string
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
	resp := &Response{}
	opts := worktree.MergeBackOptions{
		SourcePath: req.SourcePath,
		TargetPath: req.TargetPath,
		DryRun:     false,
		Remove:     false,
		TmpDir:     req.TmpDir,
		StashLabel: req.StashLabel,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			if len(plan.Commands) > 0 {
				resp.ObservedTmpPath = plan.Commands[0].Dir
			}
			return true, nil
		},
	}
	result, err := worktree.MergeBack(opts)
	if err != nil {
		return resp, err
	}
	if result == nil {
		return resp, nil
	}
	resp.Action = result.Action
	resp.Branch = result.Branch
	resp.Relation = result.Relation
	return resp, nil
}
```
