## Expected
- Exit code is 0.
- Output contains `dry-run: would rebase`.
- The history still has the old key (not rebased).

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
	assertContains(t, resp.Output, "dry-run: would rebase")
	oldDir := filepath.Join(req.WorkRoot, "oldbase")
	assertHistoryChain(t, req.ConfigHome, oldDir, oldDir)
}
```
