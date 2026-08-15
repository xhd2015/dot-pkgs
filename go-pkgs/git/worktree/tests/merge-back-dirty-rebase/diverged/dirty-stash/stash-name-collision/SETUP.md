# Scenario

**Feature**: pre-existing stash named "wrk-merge-back" doesn't cause contamination

```
# User has an old stash from a previous failed merge-back. Current run has
# different dirty changes. The old stash must not be consumed or merged in.
dirty feat -> stash push (new) -> old stash exists -> stash apply (new) -> flow succeeds -> old stash intact
```

## Steps

1. Make a dirty change, stash it as "wrk-merge-back" (simulating previous failed run).
2. Make a DIFFERENT dirty change (the current run's work).
3. MergeBack should succeed, old stash survives, new changes are correct.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := req.MainRepo

	// Diverge with main on a different file (no conflict)
	if err := os.WriteFile(filepath.Join(mainRepo, "main-only.txt"), []byte("diverging\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "main-only.txt")
	runGit(t, mainRepo, "commit", "-m", "add main-only on main")

	// Make a dirty change and stash it (simulating previous failed run)
	if err := os.WriteFile(filepath.Join(req.SourcePath, "old-dirty.txt"), []byte("old stash content\n"), 0644); err != nil {
		return err
	}
	runGit(t, req.SourcePath, "stash", "push", "-u", "-m", "wrk-merge-back")

	// Make a DIFFERENT dirty change (this is the current run's work)
	if err := os.WriteFile(filepath.Join(req.SourcePath, "fresh-dirty.txt"), []byte("fresh content\n"), 0644); err != nil {
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
