# Scenario

**Feature**: `--list-worktrees` discovers main and linked worktree rows

```
main + feature-a worktree -> two lines: main\tmain and feature-a\tworktree
```

## Steps

1. Init main repo with initial commit.
2. Add linked worktree `feature-a`.
3. Set `req.Args` to `["--root", <workspace>, "--list-worktrees"]`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	wtDir := filepath.Join(root, "feature-a")
	gitInitRepo(t, mainDir)
	gitInitialCommit(t, mainDir)
	gitWorktreeAdd(t, mainDir, wtDir, "feature-a")
	req.Args = []string{"--root", root, "--list-worktrees"}
	return nil
}
```