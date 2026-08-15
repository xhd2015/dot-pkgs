# Scenario

**Feature**: user modified 3 files, rebase modified 1 of them — all rejected

```
# Feature commits don't touch README.md. Main modifies it. Rebase clean.
# User dirties README.md (conflict) + 2 other files (clean).
# stash apply conflicts on README.md → ALL changes rejected, source untouched.
dirty feat -> stash push (3 files) -> rebase clean -> stash apply conflict on 1 -> ALL rejected
```

## Steps

1. Main modifies README.md (feature didn't touch it → rebase clean).
2. User modifies README.md + 2 other files in working tree.
3. MergeBack → stash apply conflicts on README.md → all rejected.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := req.MainRepo

	// Main: modify README.md (feature didn't touch it, rebase clean)
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("# MAIN CHANGE\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "modify README on main")

	// User dirties README.md + 2 other files
	if err := os.WriteFile(filepath.Join(req.SourcePath, "README.md"), []byte("# USER CONFLICT\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(req.SourcePath, "other-1.txt"), []byte("USER other 1\n"), 0644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(req.SourcePath, "other-2.txt"), []byte("USER other 2\n"), 0644); err != nil {
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
