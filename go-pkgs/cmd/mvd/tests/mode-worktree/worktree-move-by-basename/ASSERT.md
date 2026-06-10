## Expected
- Exit code 0.
- Output contains "worktree created:".
- History records the worktree chain with worktree metadata.

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
	assertContains(t, resp.Output, "worktree created:")

	mainRepo := filepath.Join(req.WorkRoot, "projects", "myrepo")
	wtDir := filepath.Join(req.WorkRoot, "feature")
	assertHistoryChain(t, req.ConfigHome, mainRepo, mainRepo, wtDir)
	assertHistoryWorktreeEntry(t, req.ConfigHome, mainRepo, 1, mainRepo, "feature")
}
```
