# Scenario

**Feature**: when `ListWorktrees=false`, there is no worktree expand; top-level
under-root filter still applies. Neighbor checkouts outside the scan root
must not appear; `Worktrees` stays empty/nil.

```
# layout
parent/
  A/   # scan root — main (fake .git); cold-seeded / warm-eligible
  B/   # sibling main outside A — sibling probe may ReadDir(parent)

Scan(Roots=[A], ListWorktrees=false, NoCache=false)
  -> Result paths ⊆ under A
  -> B omitted
  -> every row Worktrees empty/nil (no expand)
```

## Preconditions

- Fake `.git` only; no git CLI.
- Contrast: `list-worktrees-outside-base-stripped` exercises strip on
  `Worktrees` when the flag is true. This leaf proves flag-off still
  filters top-level `Repos` and never fills `Worktrees`.
- Cold-seed against A so warm + sibling probe can see B.
- Classic TDD: RED if sibling B leaks into Result without under-root filter.

## Steps

1. Create `parent/A` and `parent/B` with fake `.git`.
2. Cold-seed with `Roots: [A]`.
3. Set warm Run roots to A; force `ListWorktrees=false`.
4. Stash `ConsumerPath=A`, `SiblingPath=B`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	parent := t.TempDir()
	a := filepath.Join(parent, "A")
	b := filepath.Join(parent, "B")
	mkdirAll(t, a)
	mkdirAll(t, b)
	fakeGitRepo(t, a)
	fakeGitRepo(t, b)

	aAbs := absPath(t, a)
	bAbs := absPath(t, b)

	req.Roots = []string{aAbs}
	req.NoCache = false
	req.Refresh = false
	req.ListWorktrees = false
	req.ListRemotes = false
	// Soft-isolate: no unit rewalk so only index/sibling/filter matter.
	req.WarmRefreshBudget = -1

	coldSeedScan(t, req.Roots, req.CacheRoot)

	req.ConsumerPath = aAbs
	req.SiblingPath = bAbs
	req.ForeignPath = bAbs
	return nil
}
```
