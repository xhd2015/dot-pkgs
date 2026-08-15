## Expected
- Exit code 0 (default auto-yes; non-TTY no longer hard-requires confirm).
- Output contains `worktree removed:`.
- Output does **not** contain `Proceed?`.
- Worktree directory no longer exists.
- Main repo has both the feature-work and main-work files (rebase + ff merge succeeded).
- History is nil.

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

	assertNotContains(t, resp.Output, "Proceed?")
	assertContains(t, resp.Output, "worktree removed:")
	assertFileNotExists(t, wtDir)

	// Both files should exist in main after rebase + ff merge
	assertFileExists(t, filepath.Join(mainRepo, "feature-work"))
	assertFileExists(t, filepath.Join(mainRepo, "main-work"))

	assertHistoryNil(t, req.ConfigHome)
}
```
