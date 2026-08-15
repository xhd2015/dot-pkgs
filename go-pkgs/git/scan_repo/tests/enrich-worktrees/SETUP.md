# Scenario

**Feature**: `ListWorktrees=true` enriches main repos with worktree list via git

```
Scan -> git worktree list --porcelain -> Worktrees[] on RepoTypeMain only
```

## Preconditions

- Real `git` on PATH (tests skip otherwise).
- `ListWorktrees=true`.

## Steps

1. Enable `ListWorktrees`.
2. Provide git worktree helpers.

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
	req.ListWorktrees = true
	req.ListRemotes = false
	return nil
}

func gitInitRepo(t *testing.T, dir string) {
	t.Helper()
	mkdirAll(t, dir)
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "Test User")
}

func gitInitialCommit(t *testing.T, dir string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "README"), "init\n")
	runGit(t, dir, "add", "README")
	runGit(t, dir, "commit", "-m", "init")
}

func gitWorktreeAdd(t *testing.T, mainDir, wtDir, branch string) {
	t.Helper()
	mkdirAll(t, filepath.Join(mainDir, ".git")) // ensure parent exists
	runGit(t, mainDir, "worktree", "add", "-b", branch, wtDir)
}
```