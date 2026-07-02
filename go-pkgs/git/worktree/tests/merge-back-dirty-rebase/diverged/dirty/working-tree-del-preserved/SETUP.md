# Scenario

**Feature**: diverged + dirty tracked-file deletion + no-rm → deletion preserved after sync

```
# User deleted delete-me.txt (uncommitted). Rebased commit still has the file.
# After reset --mixed HEAD, index has the file, working tree does not.
dirty feat -> tmp rebase -> merge -> reset --mixed -> user deletion preserved
```

## Steps

1. Create `delete-me.txt` on feature branch and commit it.
2. Diverge with a main commit touching a **different** file (no conflict).
3. Delete `delete-me.txt` from working tree (tracked-file deletion, uncommitted).
4. Set `WRK_HOME` to temp dir.
5. Call MergeBack.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := req.MainRepo

	// Create a tracked file on feature branch
	deleteMe := filepath.Join(req.SourcePath, "delete-me.txt")
	if err := os.WriteFile(deleteMe, []byte("will be deleted\n"), 0644); err != nil {
		return err
	}
	runGit(t, req.SourcePath, "add", "delete-me.txt")
	runGit(t, req.SourcePath, "commit", "-m", "add delete-me.txt")

	// Diverge: commit on main touching a different file (no rebase conflict)
	if err := os.WriteFile(filepath.Join(mainRepo, "another-file.txt"), []byte("diverging\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "another-file.txt")
	runGit(t, mainRepo, "commit", "-m", "another file on main")

	// Delete from working tree (tracked-file deletion, NOT staged)
	if err := os.Remove(deleteMe); err != nil {
		return err
	}

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	t.Setenv("WRK_HOME", wrkHome)
	return nil
}
```
