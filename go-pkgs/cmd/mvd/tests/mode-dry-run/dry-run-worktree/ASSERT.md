## Expected
- Exit code is 0.
- Output contains `dry-run: would create worktree`.
- The worktree directory was NOT actually created.
- No history was written.

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
	assertContains(t, resp.Output, "dry-run: would create worktree")
	wtDir := filepath.Join(req.WorkRoot, "feature")
	assertFileNotExists(t, wtDir)
	assertHistoryNil(t, req.ConfigHome)
}
```
