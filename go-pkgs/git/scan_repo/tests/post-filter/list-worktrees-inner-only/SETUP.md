# Scenario

**Feature**: with `ListWorktrees=true`, linked worktrees under the scan root
appear on the owning main’s `Worktrees` (inner field); not invent-promoted
as enrich-only top-level `Repos`.

```
# layout
scanRoot/
  main/          # real git main
  main-wt/       # git worktree add under same parent

Scan(Roots=[scanRoot], ListWorktrees=true)
  -> main is top-level RepoTypeMain
  -> abs(main-wt) listed in main.Worktrees (IsMain=false)
  -> main itself in Worktrees with IsMain=true
  -> FS walk may also emit main-wt as a top-level worktree row (Option A dual
     discovery under root) — allowed; assert under-root + inner field
```

## Preconditions

- Real `git` on PATH (skip otherwise).
- Both main and linked worktree paths are under the scan root.
- Explicit temp `CacheRoot`; discovery-only cold path is fine (`NoCache` left
  false is OK — first Scan is cold full walk).
- Classic TDD: RED if ListWorktrees fails to attach under-root worktrees, or
  if product later strips under-root worktrees incorrectly.

## Steps

1. Init `main` with an initial commit.
2. `git worktree add` linked path `main-wt` under the same parent.
3. Set `Roots` to parent; enable `ListWorktrees`.
4. Stash `MainPath` / `WorktreePath`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	wtDir := filepath.Join(root, "main-wt")
	gitInitRepo(t, mainDir)
	gitInitialCommit(t, mainDir)
	gitWorktreeAdd(t, mainDir, wtDir, "main-wt")

	rootAbs := absPath(t, root)
	req.Roots = []string{rootAbs}
	req.ListWorktrees = true
	req.ListRemotes = false
	req.NoCache = true // discovery + enrich only; no warm/index side noise
	req.MainPath = absPath(t, mainDir)
	req.WorktreePath = absPath(t, wtDir)
	req.ConsumerPath = rootAbs
	return nil
}
```
