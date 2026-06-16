# Scenario

Plain move of the main repo after creating a worktree. Was a bug: used to move the worktree instead.

mvd -w repo wt → [(repo), (wt w:wt)]
mvd repo dst → [(repo), (wt w:wt), (dst)]

## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd -w repo wt` to create a worktree at work/wt.
- Use `mvd repo dst` (plain move, no -w) to move the main repo to work/dst.

The plain move should move the main repo directory (repo), NOT the worktree (wt). This is the core bug: before the fix, the plain move would incorrectly resolve to the worktree location and move wt instead of repo.

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

	// Step 2: plain move (no -w) using full path
	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{repo, dst}
	return nil
}
```
