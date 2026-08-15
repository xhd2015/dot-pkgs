> Plain move nests the main repo inside its own worktree directory.
> Not a bug — verifies the worktree's `.git` is correctly updated to point
> to the new (nested) main repo location.
# Scenario

Plain move nests the main repo inside its own worktree directory.

mvd -w repo wt → [(repo), (wt w:wt)]
mvd repo wt/sub → [(repo), (wt w:wt), (wt/sub)]  (nested)

## Steps
- Create a git repo at work/repo.
- `mvd -w repo wt` to create a worktree at work/wt.
- `mvd repo wt/sub` to plain move the main repo INTO the worktree directory.

When moveDir runs on repo → wt/sub, isGitRepo(repo) is true, so it lists
worktrees. The worktree wt is found and its .git file is updated to point
to the new main repo location (wt/sub). The main repo is now nested inside
its own worktree directory.

This is a valid (if unusual) filesystem state. The test verifies that mvd
correctly moves the repo and updates the worktree .git.

```go
import (
	"fmt"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	skipIfNoGit(t)

	repo := filepath.Join(req.WorkRoot, "repo")
	mkdirAll(t, repo)
	initGitRepo(t, repo)

	// Step 1: create worktree
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", repo, wt}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 2: plain move repo INTO the worktree dir
	subDir := filepath.Join(wt, "sub")
	req.Args = []string{repo, subDir}
	return nil
}
```
