# Scenario

**Feature**: two linked worktrees on the same branch via go-pkgs WorktreesOnBranch

```
# two linked checkouts of feature (main stays on master)
worktree.WorktreesOnBranch(feature) -> len=2
```

## Preconditions

- Git supports `git worktree add --force <path> <existing-branch>`.
- go-pkgs exposes `WorktreesOnBranch(repoPath, branch string) ([]Entry, error)`.

## Steps

1. Init main on `master`.
2. Add linked worktree with new branch `feature`.
3. Add second linked worktree on same `feature` with `--force`.
4. Query go-pkgs `WorktreesOnBranch(feature)`.

## Context

- Classic TDD: may be compile-RED until implementer re-exports WorktreesOnBranch.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	main := initRepo(t)
	linked1 := addLinkedBranch(t, main, "feature")
	linked2 := addLinkedExistingBranch(t, main, "feature", true)
	req.Op = "worktrees_on_branch"
	req.Dir = main
	req.Branch = "feature"
	req.MainPath = main
	req.LinkedPath = linked1
	req.LinkedPath2 = linked2
	return nil
}
```
