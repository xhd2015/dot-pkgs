# Scenario

**Feature**: main and linked worktree appear as separate rows (Option A)

```
main/.git dir + feature-a/.git gitlink -> two rows, same GitDir, different RepoType
```

## Steps

1. Create sibling `main/` and `feature-a/` via `fakeGitWorktree`.
2. Set `req.Roots` to workspace.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	wtDir := filepath.Join(root, "feature-a")
	mkdirAll(t, wtDir)
	fakeGitWorktree(t, mainDir, wtDir)
	req.Roots = []string{root}
	return nil
}
```