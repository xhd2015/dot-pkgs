## Expected
- After two successful back operations, the third `--back` call reports "nothing to move back" because the project is already at its original root `src`.
- The history chain contains only a single entry: `src` (the project is back at its origin).

## Exit Code
- 0 (success, no-op is not an error)

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp == nil {
        t.Fatalf("expected response, got error: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code: %d, output:\n%s", resp.ExitCode, resp.Output)
    }
    assertContains(t, resp.Output, "nothing to move back")
    
    src := filepath.Join(req.WorkRoot, "src")
    assertHistoryChain(t, req.ConfigHome, src, src)
}
```
