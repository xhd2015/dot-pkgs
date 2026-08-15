## Expected
- Non-zero exit code.
- After basename and alias resolution fail, the bare name falls through to path resolution.
- The resolved path is not a git repo, resulting in a "not a git repository" error.
- No history recorded.

## Exit Code
- Non-zero

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if resp == nil {
		assertErrIsNil(t, err)
		return
	}
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0\noutput: %s", resp.Output)
	}
	assertContains(t, resp.Output, "not a git repository")
	assertHistoryNil(t, req.ConfigHome)
}
```
