## Expected
- Non-zero exit code.
- Output contains "uncommitted changes".
- Worktree directory still exists.
- History still has the worktree entry.

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
	assertContains(t, resp.Output, "uncommitted changes")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	assertFileExists(t, wtDir)

	mainRepo := filepath.Join(req.WorkRoot, "main")
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```
