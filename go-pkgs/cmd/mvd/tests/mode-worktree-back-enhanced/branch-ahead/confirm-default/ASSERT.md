## Expected
- Exit code 0.
- Output contains "worktree removed:" indicating the worktree was successfully removed.
- Output contains the prompt confirming the merge (e.g., contains feature branch name).
- Worktree directory no longer exists.
- Main repo has the feature commit (fast-forward merged).
- History is nil (entry fully removed).

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
	assertFileExists(t, filepath.Join(mainRepo, "feature-work"))
	assertHistoryNil(t, req.ConfigHome)
}
```
