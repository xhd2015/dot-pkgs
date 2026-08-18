# Scenario

**Feature**: diverged + dirty + no-rm → tmp worktree rebase conflicts

```
# both main and feature modify same line → rebase conflict in tmp → abort, cleanup, error
dirty feat -> tmp worktree -> rebase conflicts -> abort -> cleanup -> error
```

## Steps

1. Create a conflicting change: both main and feature modify the same line of README.md.
2. Set `WRK_HOME` to temp dir.
3. Call MergeBack.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// create a conflicting change on feature branch
	conflictFile := filepath.Join(req.SourcePath, "shared.txt")
	err := os.WriteFile(conflictFile, []byte("feature version\n"), 0644)
	if err != nil {
		return err
	}
	runGit(t, req.SourcePath, "add", "shared.txt")
	runGit(t, req.SourcePath, "commit", "-m", "feature change to shared.txt")

	// create a conflicting change on main (same file, different content)
	if err := os.WriteFile(filepath.Join(req.MainRepo, "shared.txt"), []byte("main version\n"), 0644); err != nil {
		return err
	}
	runGit(t, req.MainRepo, "add", "shared.txt")
	runGit(t, req.MainRepo, "commit", "-m", "main change to shared.txt")

	wrkHome := filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(wrkHome, 0755); err != nil {
		return err
	}
	return nil
}
```
