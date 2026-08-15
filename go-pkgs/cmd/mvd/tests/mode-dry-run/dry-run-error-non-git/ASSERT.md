## Expected
- Non-zero exit code.
- Output contains `not a git repository`.
- Output does NOT contain `dry-run:`.

## Exit Code
- Non-zero

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0: %s", resp.Output)
	}
	assertContains(t, resp.Output, "not a git repository")
	assertNotContains(t, resp.Output, "dry-run:")
}
```
