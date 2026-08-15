# Scenario

Marker bug: plain move of worktree loses worktree marker in picker output.

mvd -w repo wt → [(repo), (wt, worktree)]
mvd wt moved-wt → [(repo), (wt, worktree, dead), (moved-wt, plain, NO git metadata)]

BUG: moved-wt IS a worktree on disk (the .git file was moved with it),
but mvd records it as a plain LocationEntry without Git metadata.
As a result, --picker-list shows (external main) instead of (worktree).

## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd -w repo wt` to create a worktree at work/wt.
- Use `mvd wt moved-wt` to move the worktree via plain move.
- Run `mvd --picker-list` to check the marker on moved-wt.

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

	// Step 2: plain move of the worktree
	movedWt := filepath.Join(req.WorkRoot, "moved-wt")
	req.Args = []string{wt, movedWt}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move worktree failed: %s", resp.Output)
	}

	// Step 3: run --picker-list to check markers
	req.Args = []string{"--picker-list"}
	return nil
}
```
