# Scenario

**Feature**: user modified tracked file, rebase deleted that file — modify/delete conflict

```
# "deleted by them": user modified README.md (exists on both branches), rebase deleted it.
# stash apply on tmp detects modify/delete conflict.
dirty feat -> stash push -> rebase deletes file -> stash apply -> modify/delete -> reject
```

## Steps

1. On main: delete README.md (which exists on both branches from init).
2. On feature's working tree: modify README.md.
3. MergeBack should reject, source unchanged.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	mainRepo := req.MainRepo

	// On main: delete README.md (diverges with feature which still has it)
	runGit(t, mainRepo, "rm", "README.md")
	runGit(t, mainRepo, "commit", "-m", "delete README.md on main")

	// Modify README.md in working tree (dirty)
	if err := os.WriteFile(filepath.Join(req.SourcePath, "README.md"), []byte("USER MODIFIED\n"), 0644); err != nil {
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
