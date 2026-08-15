# Scenario

**Feature**: cold Scan indexes gitlink worktree with repo_type worktree

```
main/ + feature-a/ (.git gitlink)
  -> Scan
  -> index has feature-a repo_type=worktree, git_dir=main/.git
  -> index has main as main
```

## Steps

1. Create sibling `main/` and `feature-a/` via `fakeGitWorktree`.
2. Set `req.Roots` to the workspace.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	wtDir := filepath.Join(root, "feature-a")
	mkdirAll(t, wtDir)
	fakeGitWorktree(t, mainDir, wtDir)
	req.Roots = []string{root}
	return nil
}
```
