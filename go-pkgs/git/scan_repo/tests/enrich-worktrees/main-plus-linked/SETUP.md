# Scenario

**Feature**: main row lists all worktrees; worktree row has empty Worktrees

```
main + feature-a linked -> 2 scan rows; Worktrees only on main row
```

## Steps

1. Init main repo with commit.
2. Add linked worktree `feature-a`.
3. Set `req.Roots` to workspace containing both checkouts.

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
	wtDir := filepath.Join(root, "feature-a")
	gitInitRepo(t, mainDir)
	gitInitialCommit(t, mainDir)
	gitWorktreeAdd(t, mainDir, wtDir, "feature-a")
	req.Roots = []string{root}
	return nil
}
```