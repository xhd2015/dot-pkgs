# Scenario

**Feature**: `ListWorktrees=false` skips git subprocesses (fake repo succeeds)

```
ListWorktrees=false on fake .git -> scan succeeds without git binary
```

## Steps

1. Override `ListWorktrees` to false (parent branch sets true).
2. Create fake main + worktree via scan helpers pattern.
3. Set `req.Roots` to workspace.

```go
import (
	"path/filepath"
	"testing"
)

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}

func fakeGitWorktree(t *testing.T, mainDir, wtDir string) {
	t.Helper()
	fakeGitRepo(t, mainDir)
	wtName := filepath.Base(wtDir)
	wtGitDir := filepath.Join(mainDir, ".git", "worktrees", wtName)
	mkdirAll(t, wtGitDir)
	absWtGitDir := absPath(t, wtGitDir)
	writeFile(t, filepath.Join(wtDir, ".git"), "gitdir: "+absWtGitDir+"\n")
}

func Setup(t *testing.T, req *Request) error {
	req.ListWorktrees = false
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	wtDir := filepath.Join(root, "feature-a")
	mkdirAll(t, wtDir)
	fakeGitWorktree(t, mainDir, wtDir)
	req.Roots = []string{root}
	return nil
}
```