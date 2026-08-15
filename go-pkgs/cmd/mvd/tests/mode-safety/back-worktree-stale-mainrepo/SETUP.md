> `--back` on a worktree after its main repo was plain-moved.
> **Was a bug (fixed)**: `resolveBackEntry` used to block with "position mismatch";
> now allows worktree entries at any position in the chain. MainRepo is read
> from the worktree's `.git` file on disk instead of stale history.
# Scenario

--back on a worktree after its main repo was plain-moved. Was a bug: blocked by position mismatch; now reads MainRepo from .git file on disk.

mvd -w repo wt → [(repo), (wt w:wt)]
mvd repo mid → [(repo), (wt w:wt), (mid)]
mvd --back wt → [(repo), (mid)]  (wt removed, mid preserved)

## Steps
- Create a git repo at work/repo.
- `mvd -w repo wt` to create a worktree (MainRepo = repo).
- `mvd repo mid` to plain move the main repo to work/mid.
  moveDir correctly updates wt/.git to point to mid.
- `mvd --back wt` to attempt back on the worktree.

After the plain move, the history chain is [repo, wt(worktree), mid].
The worktree wt is NOT the last entry. The current code checks that the
resolved path must be root or last in the chain. Since wt is neither,
--back fails with "current position mismatch".

This is a BUG because:
1. The user should be able to --back a worktree even after the main repo moved.
2. If the position check were relaxed, cmdWorktreeBack would read
   last.Git.MainRepo = "repo" from history — a stale value. The main repo was
   moved to "mid" and "repo" no longer exists on disk.
   The fix should read the MainRepo from the worktree's .git file instead
   of trusting history.

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

	// Step 2: plain move of main repo — moveDir updates wt/.git correctly
	mid := filepath.Join(req.WorkRoot, "mid")
	req.Args = []string{repo, mid}
	resp, err = runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move repo→mid failed: %s", resp.Output)
	}

	// Step 3: attempt --back on worktree
	req.Args = []string{"--back", wt}
	return nil
}
```
