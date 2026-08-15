## Expected
- Exit code 0.
- Output contains "worktree removed:" and "branch: feature deleted".
- Worktree directory no longer exists.
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
	assertContains(t, resp.Output, "worktree removed:")
	assertContains(t, resp.Output, "branch: feature deleted")

	wtDir := filepath.Join(req.WorkRoot, "feature")
	assertFileNotExists(t, wtDir)
	assertHistoryNil(t, req.ConfigHome)
}
```
