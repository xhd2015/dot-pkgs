## Expected
- Exit code 0 (no error — operation aborted cleanly).
- Output indicates the operation was cancelled/aborted.
- Worktree directory still exists.
- Feature branch still exists.
- Main repo does NOT have the feature commit.
- History unchanged (worktree entry still present).

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")

	// Worktree still exists
	assertFileExists(t, wtDir)
	assertFileExists(t, filepath.Join(wtDir, ".git"))

	// Main does NOT have the feature commit
	assertFileNotExists(t, filepath.Join(mainRepo, "feature-work"))

	// History still has the worktree entry
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```
