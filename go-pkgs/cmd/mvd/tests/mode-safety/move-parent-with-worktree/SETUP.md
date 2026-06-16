> Moves a parent directory containing a tracked repo and its worktree.
> **Was a bug (fixed)**: the worktree .git file used to stay stale after the move;
> now `moveDir` recursively discovers and updates it.
# Scenario

Move a parent directory containing a tracked repo and its worktree. Was a bug: worktree .git stayed stale; now recursively updated.

mvd --add parent/repo → [(parent/repo)]
mvd -w parent/repo parent/wt → [(parent/repo), (parent/wt w:parent/wt)]
mvd parent elsewhere → [parent/repo entry dead], [(parent), (elsewhere)]  (wt .git updated)

## Steps
- Create a parent directory containing a git repo: parent/repo.
- `mvd --add parent/repo` to track the repo.
- `mvd -w parent/repo parent/wt` to create a worktree inside the parent.
- `mvd parent elsewhere` to move the entire parent directory.

The parent dir move is a plain move. Since `isGitRepo(parent)` is false (the .git
is inside parent/repo/.git, not parent/.git), `moveDir` does NOT list worktrees
and does NOT update the worktree .git file. The worktree .git stays stale,
referencing the old parent/repo path. History for parent/repo still shows dead
paths.

This is a BUG: the worktree .git file should be updated when the parent containing
the main repo is moved. Currently it is not.

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

	// Step 1: add parent/repo to tracking
	req.Args = []string{"--add", repo}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("--add parent/repo failed: %s", resp.Output)
	}

	// Step 2: create worktree inside parent
	wt := filepath.Join(parent, "wt")
	req.Args = []string{"-w", repo, wt}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 3: move parent dir → elsewhere
	elsewhere := filepath.Join(req.WorkRoot, "elsewhere")
	req.Args = []string{parent, elsewhere}
	return nil
}
```
