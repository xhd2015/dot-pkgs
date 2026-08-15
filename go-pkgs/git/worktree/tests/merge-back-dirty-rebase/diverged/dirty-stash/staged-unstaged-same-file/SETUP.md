# Scenario

**Feature**: user has staged and unstaged changes on same file — both survive

```
# User staged line 2 change, then made unstaged line 3 change on same file.
# Rebase touches a different file. Stash captures both, re-applies both.
dirty feat -> stash push -> rebase -> stash apply clean -> migrate -> both changes survive
```

## Steps

1. Create `multi.txt` on feature and commit.
2. On main: modify a different file (diverges, no conflict with multi.txt).
3. Stage a change to `multi.txt` line 2, then make an unstaged change to line 3.
4. MergeBack should succeed, both changes preserved (though stage info may be lost).

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	mainRepo := req.MainRepo

	// Create and commit multi.txt with 3 lines
	multiFile := filepath.Join(req.SourcePath, "multi.txt")
	content := "line 1\nline 2\nline 3\n"
	if err := os.WriteFile(multiFile, []byte(content), 0644); err != nil {
		return err
	}
	runGit(t, req.SourcePath, "add", "multi.txt")
	runGit(t, req.SourcePath, "commit", "-m", "add multi.txt on feature")

	// Diverge: modify different file on main (no conflict with multi.txt)
	if err := os.WriteFile(filepath.Join(mainRepo, "main-only.txt"), []byte("diverging\n"), 0644); err != nil {
		return err
	}
	runGit(t, mainRepo, "add", "main-only.txt")
	runGit(t, mainRepo, "commit", "-m", "add main-only.txt on main")

	// Stage change to line 2
	if err := os.WriteFile(multiFile, []byte("line 1\nLINE TWO STAGED\nline 3\n"), 0644); err != nil {
		return err
	}
	runGit(t, req.SourcePath, "add", "multi.txt")

	// Unstaged change to line 3
	if err := os.WriteFile(multiFile, []byte("line 1\nLINE TWO STAGED\nLINE THREE UNSTAGED\n"), 0644); err != nil {
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
