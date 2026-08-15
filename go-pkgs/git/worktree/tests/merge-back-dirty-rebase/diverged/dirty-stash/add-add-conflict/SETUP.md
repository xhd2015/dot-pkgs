# Scenario

**Feature**: user has untracked file, rebase creates same file — add/add conflict

```
# user created new.txt (untracked), rebase also created new.txt with different content.
# stash push -u captures it, stash apply on tmp conflicts.
dirty feat -> stash push -u -> rebase creates same file -> stash apply -> add/add conflict -> reject
```

## Steps

1. Create untracked `new.txt` on feature's working tree.
2. On main: create and commit `new.txt` with different content.
3. MergeBack should reject, source unchanged.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := req.MainRepo

	// Create untracked new.txt on feature (NOT staged, NOT committed)
	newFile := filepath.Join(req.SourcePath, "new.txt")
	if err := os.WriteFile(newFile, []byte("USER untracked\n"), 0644); err != nil {
		return err
	}

	// On main: create new.txt with different content (diverges)
	if err := os.WriteFile(filepath.Join(mainRepo, "new.txt"), []byte("MAIN created\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "new.txt")
	runGit(t, mainRepo, "commit", "-m", "add new.txt on main")

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	t.Setenv("WRK_HOME", wrkHome)
	return nil
}
```
