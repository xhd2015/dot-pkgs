# Scenario

**Feature**: gitlink `.git` file classifies as worktree

```
wtDir/.git file (gitdir:) -> RepoTypeWorktree, GitDir resolves to main/.git
```

## Steps

1. Create main repo and linked worktree via `fakeGitWorktree`.
2. Scan only the worktree checkout path as root child.

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