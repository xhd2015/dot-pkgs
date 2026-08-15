## Expected
- The rebase command exits with code 0.
- The output contains "rebased:".
- The history chain keyed by newBase has paths [newBase, dir].

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
    newBase := filepath.Join(req.WorkRoot, "newbase")
    dir := filepath.Join(req.WorkRoot, "projects", "myproject")
    assertHistoryChain(t, req.ConfigHome, newBase, newBase, dir)
}
```
