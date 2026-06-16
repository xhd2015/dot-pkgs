> `--back` from a repo nested inside its own worktree directory.
> Not a bug — verifies the skip-worktree-in-prev logic and correct `.git` update.
# Scenario

--back from a repo nested inside its own worktree. Verifies skip-worktree-in-prev logic.

(replay move-into-worktree-dir) → [(repo), (wt w:wt), (wt/sub)]
mvd --back wt/sub → [(repo), (wt w:wt)]

## Steps
- Replay Scenario D as setup: create repo, worktree at wt, move repo into wt/sub.
- Then `mvd --back wt/sub` to restore the main repo back to its original location.

The --back should:
1. Find prev by skipping the worktree entry (wt).
2. Move wt/sub back to repo.
3. Update the worktree .git to point back to repo.

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

	// Step 2: move repo INTO worktree dir
	subDir := filepath.Join(wt, "sub")
	req.Args = []string{repo, subDir}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move repo→wt/sub failed: %s", resp.Output)
	}

	// Step 3: --back to restore from nested location
	req.Args = []string{"--back", subDir}
	return nil
}
```
