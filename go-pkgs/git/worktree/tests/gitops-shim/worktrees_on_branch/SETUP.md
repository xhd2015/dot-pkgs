# Scenario

**Feature**: WorktreesOnBranch via go-pkgs filters registered entries by branch

```
# fixture with two linked worktrees on same branch
go-pkgs worktree.WorktreesOnBranch(repo, branch) -> Entry slice (no policy error)
```

## Preconditions

- Root helpers for init / linked worktrees.
- `WorktreesOnBranch` is part of go-pkgs surface (re-export from gitops if needed).

## Steps

1. Grouping only; leaf configures branch and dual linked fixtures.

## Context

- Scenario 2 from P2: multi-checkout is data only at this layer (refuse is P3).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// Grouping: leaf sets Branch and dual linked paths.
	return nil
}
```
