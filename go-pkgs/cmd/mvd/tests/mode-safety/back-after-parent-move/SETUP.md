> After moving a parent that contained a tracked repo + worktree, `--back`
> restores the parent and cleans up orphaned sub-project history entries.
> **Was a bug (fixed)**: sub-project entries used to remain dead; now auto-cleaned.
# Scenario

--back after parent move restores parent and cleans up orphaned sub-project entries. Was a bug: sub-entries remained dead.

(replay move-parent-with-worktree) → [(parent/repo), (parent/wt w:parent/wt)], [(parent), (elsewhere)]
mvd --back elsewhere → [(parent)]  (sub-entry cleaned up)

## Steps
- Replay Scenario A (move-parent-with-worktree) as setup.
- Then `mvd --back elsewhere` to restore the parent directory.

The --back on elsewhere resolves to the parent dir's own history entry
and moves elsewhere back to parent. However, the parent/repo history entry
(with its worktree) is NOT touched — those paths remain dead.

After the back, parent/ exists again with repo/ and wt/ inside, but the
worktree .git was never updated during the initial parent move (Scenario A
bug), so it still points to the old parent/repo path. Since parent/repo
now exists again physically, the stale reference happens to resolve, but
it was wrong during the time parent/ was at elsewhere/.

```go
import (
	"fmt"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	parent := filepath.Join(req.WorkRoot, "parent")
	repo := filepath.Join(parent, "repo")
	mkdirAll(t, repo)
	initGitRepo(t, repo)

	// Step 1: add and create worktree (same as Scenario A)
	req.Args = []string{"--add", repo}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("--add failed: %s", resp.Output)
	}

	wt := filepath.Join(parent, "wt")
	req.Args = []string{"-w", repo, wt}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 2: move parent → elsewhere
	elsewhere := filepath.Join(req.WorkRoot, "elsewhere")
	req.Args = []string{parent, elsewhere}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("parent move failed: %s", resp.Output)
	}

	// Step 3: --back elsewhere → should restore to parent
	req.Args = []string{"--back", elsewhere}
	return nil
}
```
