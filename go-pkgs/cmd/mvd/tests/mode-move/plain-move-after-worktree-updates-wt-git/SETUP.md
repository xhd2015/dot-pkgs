## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd -w repo wt` to create a worktree at work/wt.
- Use `mvd repo dst` (plain move) to move the main repo to work/dst.
- Verify that wt/.git is updated to point to the new main repo location at dst.

When moveDir is called on the main repo (repo → dst), it lists linked worktrees and updates each worktree's .git file. The worktree wt should have its gitdir updated from repo/.git/worktrees/... to dst/.git/worktrees/...

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

	// Step 2: plain move the main repo
	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{repo, dst}
	return nil
}
```
