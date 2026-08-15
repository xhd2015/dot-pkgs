## Expected
- Exit code 0.
- Output contains "added:".
- History records the resolved path.

## Exit Code
- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	dir := filepath.Join(req.WorkRoot, "projects", "myproject")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "added:")
	assertHistoryChain(t, req.ConfigHome, dir, dir)
}
```
