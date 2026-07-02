# Merge-back: dirty worktree with rebase — tmp worktree fallback

## Version
0.0.2

When `worktree.MergeBack()` encounters a diverged branch relation AND the
source worktree is dirty AND `Remove` is false, instead of failing with
"worktree is not clean", the operation spawns a temporary worktree for the
rebase. The source worktree's uncommitted changes are preserved.

All other cases (ahead+clean, ahead+dirty, diverged+clean, diverged+dirty+rm,
same/ancestor+any) retain their existing behavior.

The tmp worktree is created at `~/.wrk/worktrees/` following the same naming
convention as `wrk` worktrees: `<repo>-<branchToken>-<date>-tmp-rebase[-suffix]`.
Tmp branch follows the same convention: `<source-branch>-tmp-rebase-<random>`.
Both are always cleaned up (success or failure).

## DSN (Domain Specific Notion)

### Participants

- **Worktree** — a linked git worktree (`.git` is a file pointing to main
  repo's `.git/worktrees/<name>`). Has a checked-out branch, a working tree,
  and may have uncommitted (dirty) changes.
- **Main repo** — the `.git` bare-ish repository that owns worktrees.
  Contains all branches and the `worktrees/` management dir.
- **Relation** — the git topology between the worktree HEAD and the target
  HEAD: `same`, `ancestor`, `ahead`, or `diverged`.
- **MergeBack** — the orchestrator function that reads relation, decides what
  git commands to run, and executes them.
- **Tmp worktree** — a throwaway linked worktree at `~/.wrk/worktrees/` used
  only when the source worktree is dirty and a rebase is needed. Always
  cleaned up after the operation.

### Behaviors

**MergeBack decision matrix (relation × clean × Remove):**

| Relation  | Clean? | Remove? | Behavior |
|-----------|--------|---------|----------|
| diverged  | dirty  | false   | **tmp worktree**: create tmp, rebase there, merge, force-update source branch, clean up |
| diverged  | dirty  | true    | error: "worktree not clean" (existing) |
| diverged  | clean  | false   | direct rebase in source, merge (existing) |
| diverged  | clean  | true    | direct rebase in source, merge, remove (existing) |
| ahead     | dirty  | any     | error: "worktree not clean" (existing) |
| ahead     | clean  | false   | ff-merge (existing) |
| ahead     | clean  | true    | ff-merge, remove (existing) |
| same/anc  | dirty  | any     | error: "worktree not clean" (existing) |
| same/anc  | clean  | false   | noop (existing) |
| same/anc  | clean  | true    | remove only (existing) |

**Tmp worktree lifecycle (diverged + dirty + !Remove):**

1. Create a temporary branch from source HEAD
2. Create a tmp worktree on that branch at `~/.wrk/worktrees/<repo>-<pathToken>-<date>-tmp-rebase[-suffix]`
3. Run `git rebase <target-HEAD>` inside tmp worktree
4. FF-merge rebased result into target: `git merge --ff-only <tmp-branch>`
5. Force-update source branch: `git branch -f <source-branch> <tmp-branch>`
6. Remove tmp worktree: `git worktree remove <tmp-path>`
7. Delete tmp branch: `git branch -D <tmp-branch>`

On rebase conflict at step 3: `git rebase --abort` in tmp worktree, then
cleanup (steps 6–7). Source branch is NOT force-updated. Error returned.

## Decision Tree

```
merge-back-dirty-rebase
├── not-diverged/                     relation is "ahead" or "same"/"ancestor"
│   ├── ahead-dirty/                  ahead + dirty → still errors
│   └── included-dirty/               ancestor/same + dirty → still errors
└── diverged/                         relation is "diverged"
    ├── with-rm/                      Remove=true
    │   └── dirty-errors/             diverged + dirty + --rm → still errors
    ├── clean/                        clean source worktree
    │   └── direct-rebase/            diverged + clean + no-rm → direct rebase (existing)
    └── dirty/                        dirty, Remove=false → new tmp worktree path
        ├── success/                  happy path: tmp created, rebased, merged, cleaned
        ├── rebase-conflict/          rebase conflicts → abort, cleanup, error
        └── tmp-path-collision/       existing tmp path → suffix increment
```

## Test Index

| Leaf | Description |
|------|-------------|
| `not-diverged/ahead-dirty` | Ahead + dirty: error "worktree is not clean" |
| `not-diverged/included-dirty` | Already included + dirty: error "worktree is not clean" |
| `diverged/with-rm/dirty-errors` | Diverged + dirty + `--rm`: error "worktree is not clean" |
| `diverged/clean/direct-rebase` | Diverged + clean: direct rebase in source, merge into target |
| `diverged/dirty/success` | Diverged + dirty + no-rm: tmp worktree rebase, merge, cleanup, branch force-updated |
| `diverged/dirty/rebase-conflict` | Diverged + dirty + no-rm + conflicting commits: abort, cleanup, source unchanged |
| `diverged/dirty/tmp-path-collision` | Diverged + dirty + no-rm: pre-existing tmp dir → suffix -1 used |

## How to Run

```sh
doctest vet ./go-pkgs/git/worktree/tests/merge-back-dirty-rebase
doctest test ./go-pkgs/git/worktree/tests/merge-back-dirty-rebase
```

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/worktree"
)

type Request struct {
	WorkRoot   string
	MainRepo   string
	SourcePath string
	TargetPath string
	Remove     bool
	MakeDirty  bool
}

type Response struct {
	Action   string
	Branch   string
	Relation string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	opts := worktree.MergeBackOptions{
		SourcePath: req.SourcePath,
		TargetPath: req.TargetPath,
		DryRun:     false,
		Remove:     req.Remove,
		Confirm: func(plan worktree.MergeBackPlan) (bool, error) {
			return true, nil
		},
	}
	result, err := worktree.MergeBack(opts)
	if err != nil {
		return &Response{}, err
	}
	if result == nil {
		return &Response{}, nil
	}
	return &Response{
		Action:   result.Action,
		Branch:   result.Branch,
		Relation: result.Relation,
	}, nil
}
```
