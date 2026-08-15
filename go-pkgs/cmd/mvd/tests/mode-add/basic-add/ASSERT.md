## Expected
- mvd records the directory in history.
- Output contains "added:".
- History has one project with the tracked directory as both root and current location.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	dir := filepath.Join(req.WorkRoot, "tracked")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "added:")
	assertHistoryChain(t, req.ConfigHome, dir, dir)
}
```
