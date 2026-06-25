## Expected
- Exit code is 0 (validation passes — worktree is clean, branch is merged).
- Output contains planned `git -C` commands from MergeBack dry-run, including `worktree remove`.
- Output contains `dry-run: would remove worktree`.
- The worktree directory still exists.
- History still has the worktree entry.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "git -C")
	assertContains(t, resp.Output, "worktree remove")
	assertContains(t, resp.Output, "dry-run: would remove worktree")
	wtDir := filepath.Join(req.WorkRoot, "feature")
	assertFileExists(t, filepath.Join(wtDir, ".git"))
	mainRepo := filepath.Join(req.WorkRoot, "main")
	assertHistoryLen(t, req.ConfigHome, 1)
	// History still has worktree entry
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```
