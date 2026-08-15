# Scenario

**Feature**: user deleted tracked file, rebase modified it — delete/modify conflict

```
# "deleted by us": user deleted README.md from working tree, rebase modified it.
# stash apply on tmp detects delete/modify conflict.
dirty feat -> stash push -> rebase modifies -> stash apply -> delete/modify -> reject
```

## Steps

1. On main: modify README.md (diverges).
2. Delete README.md from feature's working tree.
3. MergeBack should reject, source unchanged.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := req.MainRepo

	// On main: modify README.md (diverges with feature)
	if err := os.WriteFile(filepath.Join(mainRepo, "README.md"), []byte("# MAIN MODIFIED\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "README.md")
	runGit(t, mainRepo, "commit", "-m", "modify README.md on main")

	// Delete README.md from working tree (dirty, tracked-file deletion)
	if err := os.Remove(filepath.Join(req.SourcePath, "README.md")); err != nil {
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
