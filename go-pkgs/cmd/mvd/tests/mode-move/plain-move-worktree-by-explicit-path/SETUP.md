# Scenario

Explicit worktree path still moves the worktree itself.

mvd -w repo wt → [(repo), (wt w:wt)]
mvd wt dst → [(repo), (dst w:wt)]

## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd -w repo wt` to create a worktree at work/wt.
- Use `mvd wt dst` (plain move, no -w) with the explicit worktree path as source.

When the user provides the exact worktree path, this is an explicit request to move the worktree itself (not the main repo). The plain move should honor the explicit path and move wt to dst.

```go
import (
	"fmt"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	repo := filepath.Join(req.WorkRoot, "repo")
	mkdirAll(t, repo)
	initGitRepo(t, repo)

	// Step 1: create worktree
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", repo, wt}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 2: plain move with explicit worktree path
	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{wt, dst}
	return nil
}
```
