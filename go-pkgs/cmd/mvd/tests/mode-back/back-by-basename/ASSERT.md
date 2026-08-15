## Expected
- The project is moved back to its original location at `projects/kool`.
- The output contains "moved back", confirming the back operation succeeded.
- The moved location `scratch/kool` no longer exists.
- The history chain is reduced to a single entry at `projects/kool` (back at origin).

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
    assertContains(t, resp.Output, "moved back")
    
    src := filepath.Join(req.WorkRoot, "projects", "kool")
    dstKool := filepath.Join(req.WorkRoot, "scratch", "kool")
    assertFileExists(t, src)
    assertFileNotExists(t, dstKool)
    
    assertHistoryChain(t, req.ConfigHome, src, src)
}
```
