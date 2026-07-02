# Scenario

**Feature**: source has only untracked files (no tracked dirt) — flow succeeds

```
# No tracked-file changes, only untracked files. stash push -u captures them.
# Rebase doesn't touch them. stash apply clean. Flow succeeds.
dirty feat (untracked only) -> stash push -u -> rebase -> stash apply clean -> migrate -> succeed
```

## Steps

1. Create only untracked files on feature's working tree.
2. Diverge with main on a different file.
3. MergeBack should succeed, untracked files survive.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := req.MainRepo

	// Create untracked files (NOT staged, NOT committed)
	if err := os.WriteFile(filepath.Join(req.SourcePath, "untracked-1.txt"), []byte("untracked one\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(req.SourcePath, "untracked-2.txt"), []byte("untracked two\n"), 0644); err != nil {
		return err
	}

	// Diverge: modify a different file on main (no conflict with untracked)
	if err := os.WriteFile(filepath.Join(mainRepo, "main-only.txt"), []byte("diverging\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "main-only.txt")
	runGit(t, mainRepo, "commit", "-m", "add main-only on main")

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	t.Setenv("WRK_HOME", wrkHome)
	return nil
}
```
