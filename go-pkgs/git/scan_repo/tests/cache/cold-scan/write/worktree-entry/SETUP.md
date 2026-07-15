# Scenario

**Feature**: cold Scan marks gitlink worktree in the mirror cache

```
# main + linked worktree
main/ + feature-a/ (.git gitlink)
  -> Scan
  -> LoadCacheEntry(feature-a): is_repo=true, repo_type=worktree, git_dir=main/.git
```

## Steps

1. Create sibling `main/` and `feature-a/` via `fakeGitWorktree`.
2. Set `req.Roots` to the workspace.

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
