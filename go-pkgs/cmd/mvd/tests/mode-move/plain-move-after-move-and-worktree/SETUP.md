# Scenario

After repo→mid, worktree from mid, plain move of repo moves mid (not worktree).

mvd repo mid → [(repo), (mid)]
mvd -w mid wt → [(repo), (mid), (wt w:wt)]
mvd repo dst → [(repo), (mid), (wt w:wt), (dst)]

## Steps
- Create a git repo at work/repo with one commit.
- First, `mvd repo mid` to move the main repo to work/mid.
- Then `mvd -w mid wt` to create a worktree at work/wt from work/mid.
- Then `mvd repo dst` (plain move using the original root basename "repo").

The plain move should skip the worktree and find the main repo's current location (mid), then move mid to dst. The worktree wt should remain unaffected (except its .git file should be updated to point to dst).

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

	// Step 1: move repo → mid
	mid := filepath.Join(req.WorkRoot, "mid")
	req.Args = []string{repo, mid}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("first move failed: %s", resp.Output)
	}

	// Step 2: create worktree from mid
	wt := filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-w", mid, wt}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 3: plain move using original basename "repo"
	dst := filepath.Join(req.WorkRoot, "dst")
	req.Args = []string{"repo", dst}
	return nil
}
```
