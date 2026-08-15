## Expected
- Exit code is 0.
- Output contains `dry-run: would create worktree`.
- work/wt2 was NOT created.
- No history was written.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "dry-run: would create worktree")

	wt2 := filepath.Join(req.WorkRoot, "wt2")
	assertFileNotExists(t, wt2)
	assertHistoryNil(t, req.ConfigHome)
}
```