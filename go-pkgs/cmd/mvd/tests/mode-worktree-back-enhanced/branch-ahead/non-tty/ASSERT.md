## Expected
- Non-zero exit code.
- Output indicates that interactive confirmation is required but stdin is not a TTY.
- Output does NOT say "not merged" (the new TTY check replaces the old merge check).
- Worktree directory still exists.
- Feature branch still exists.
- Main repo does NOT have the feature commit.
- History unchanged.

## Exit Code
- Non-zero

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0\noutput: %s", resp.Output)
	}

	// Should not be the old "not merged" error — the new TTY check replaces it.
	assertNotContains(t, resp.Output, "not merged")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")

	// Worktree still exists
	assertFileExists(t, wtDir)

	// Main does NOT have the feature commit
	assertFileNotExists(t, filepath.Join(mainRepo, "feature-work"))

	// History still has the worktree entry
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```
