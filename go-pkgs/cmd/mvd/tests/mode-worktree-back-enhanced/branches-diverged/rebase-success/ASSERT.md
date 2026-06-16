## Expected
- Exit code 0.
- Output contains "worktree removed:".
- Worktree directory no longer exists.
- Main repo has both the feature-work and main-work files (rebase + ff merge succeeded).
- History is nil.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")

	assertContains(t, resp.Output, "worktree removed:")
	assertFileNotExists(t, wtDir)

	// Both files should exist in main after rebase + ff merge
	assertFileExists(t, filepath.Join(mainRepo, "feature-work"))
	assertFileExists(t, filepath.Join(mainRepo, "main-work"))

	assertHistoryNil(t, req.ConfigHome)
}
```
