## Expected
- Exit code 0 (default auto-yes; non-TTY ahead no longer requires confirm).
- Output contains `worktree removed:`.
- Output does **not** contain `Proceed?`.
- Output does **not** say `not merged`.
- Worktree directory no longer exists.
- Main repo has the feature commit.
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

	assertNotContains(t, resp.Output, "not merged")
	assertNotContains(t, resp.Output, "Proceed?")
	assertContains(t, resp.Output, "worktree removed:")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	mainRepo := filepath.Join(req.WorkRoot, "main")

	assertFileNotExists(t, wtDir)
	assertFileExists(t, filepath.Join(mainRepo, "feature-work"))
	assertHistoryNil(t, req.ConfigHome)
}
```
