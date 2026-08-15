# Scenario

**Feature**: diverged + dirty + no-rm → source worktree index stays in sync with new HEAD

```
# After tmp-worktree rebase, the source worktree's index must reflect the rebased HEAD,
# not the pre-rebase commit. A stale index causes spurious staged deletions/modifications.
dirty feat -> tmp worktree rebase -> merge -> force-update branch -> source index synced
```

## Steps

1. The source worktree already has committed changes (from parent SETUPs).
2. Diverge from main with a new commit on main.
3. Make source dirty.
4. Call MergeBack.
5. Assert: `git status --porcelain` in source shows NO staged (index) changes — only untracked/modified from the dirtiness. The `?? dirty.txt` should be present, but no `D ` (staged delete) or `M ` (index modified) lines.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Add a commit on feature BEFORE diverging, so that after rebase
	// the rebased commits include files that the pre-rebase index didn't have.
	if err := os.WriteFile(filepath.Join(req.SourcePath, "extra-file.txt"), []byte("extra\n"), 0644); err != nil {
		return err
	}
	runGit(t, req.SourcePath, "add", "extra-file.txt")
	runGit(t, req.SourcePath, "commit", "-m", "add extra-file on feature")

	// Now diverge: commit on main that changes README.md (shared file)
	if err := os.WriteFile(filepath.Join(req.MainRepo, "README.md"), []byte("# updated\n"), 0644); err != nil {
		return err
	}
	runGit(t, req.MainRepo, "add", "README.md")
	runGit(t, req.MainRepo, "commit", "-m", "update readme on main")

	// Make source dirty
	makeDirty(t, req.SourcePath)

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	t.Setenv("WRK_HOME", wrkHome)
	return nil
}
```
