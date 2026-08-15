# Scenario

--back on a worktree after the main repo was plain-moved. The worktree entry is removed while preserving subsequent moves.

mvd repo mid → [(repo), (mid)]
mvd -w mid wt → [(repo), (mid), (wt w:wt)]
mvd --back wt → [(repo), (mid)]  (wt removed)

## Steps
- Create a git repo at work/repo with one commit.
- Use `mvd repo mid` to move the main repo to work/mid.
- Use `mvd -w mid wt` to create a worktree from mid.
- Use `mvd --back wt` to remove the worktree.

The --back on a worktree should trigger cmdWorktreeBack: check clean status, check branch is merged, then git worktree remove + git branch -D. The history should roll back to [repo, mid].

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

	// Step 1: move repo → mid
	mid := filepath.Join(req.WorkRoot, "mid")
	req.Args = []string{repo, mid}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move repo→mid failed: %s", resp.Output)
	}

	// Step 2: create worktree from mid
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", mid, wt}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 3: --back to remove the worktree
	req.Args = []string{"--back", wt}
	return nil
}
```
