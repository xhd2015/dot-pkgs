## Expected
- The clear command exits with code 0.
- The output contains "cleared history".
- The history file is nil (no projects remain).

## Exit Code
- 0

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    assertErrIsNil(t, err)
    if resp.ExitCode != 0 {
        t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
    }
    assertContains(t, resp.Output, "cleared history")
    assertHistoryNil(t, req.ConfigHome)
}
```
