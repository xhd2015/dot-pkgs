# Scenario

**Bug / filter**: with `ListWorktrees=true`, worktree paths **outside** the
scan root must be stripped from `Repo.Worktrees` (or never attached).

```
# layout
base/
  main/              # scan root = main only (real git)
  outside/
    feature-outer/   # git worktree add — path NOT under scan root

Scan(Roots=[main], ListWorktrees=true)
  -> main appears (top-level and/or as IsMain worktree entry)
  -> feature-outer MUST NOT appear in main.Worktrees
  -> any remaining Worktrees paths ⊆ under main
```

## Preconditions

- Real `git` on PATH (skip otherwise).
- Scan root is the **main directory only**, not the parent that contains
  the outer worktree.
- Explicit temp `CacheRoot`; `NoCache=true` keeps the leaf on cold discover
  + enrich (no warm noise).
- Classic TDD: **RED** while `listWorktrees` attaches every porcelain path
  without under-root filter after resolve.

## Steps

1. Init `main` with commit under `base/`.
2. Add linked worktree at `base/outside/feature-outer`.
3. Set `Roots` to abs(`main`) only; enable `ListWorktrees`.
4. Stash `MainPath` / `WorktreePath` (outer) / `ConsumerPath` (= main).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	base := t.TempDir()
	mainDir := filepath.Join(base, "main")
	outerWt := filepath.Join(base, "outside", "feature-outer")
	gitInitRepo(t, mainDir)
	gitInitialCommit(t, mainDir)
	gitWorktreeAdd(t, mainDir, outerWt, "feature-outer")

	mainAbs := absPath(t, mainDir)
	outerAbs := absPath(t, outerWt)

	// Scan root = main only — outer worktree is outside the base.
	req.Roots = []string{mainAbs}
	req.ListWorktrees = true
	req.ListRemotes = false
	req.NoCache = true
	req.MainPath = mainAbs
	req.WorktreePath = outerAbs
	req.ForeignPath = outerAbs
	req.ConsumerPath = mainAbs
	return nil
}
```
