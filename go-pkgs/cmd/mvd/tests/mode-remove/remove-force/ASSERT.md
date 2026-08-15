## Expected
- The output contains "will clear", indicating that the force flag is removing history.
- The history file is empty after the forced removal.

## Exit Code
- 0 (success)

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp == nil {
        t.Fatalf("expected response, got error: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code: %d, output:\n%s", resp.ExitCode, resp.Output)
    }
    assertContains(t, resp.Output, "will clear")
    assertHistoryNil(t, req.ConfigHome)
}
```
