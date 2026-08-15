## Expected
- The rebase command exits with code 0.
- The output contains "rebased:".
- The history chain reflects the rebase: key is newBase, with paths [newBase, src, d1/src].

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
    assertContains(t, resp.Output, "rebased:")
    newBase := filepath.Join(req.WorkRoot, "rebased")
    src := filepath.Join(req.WorkRoot, "src")
    d1Src := filepath.Join(req.WorkRoot, "d1", "src")
    assertHistoryChain(t, req.ConfigHome, newBase, newBase, src, d1Src)
}
```
