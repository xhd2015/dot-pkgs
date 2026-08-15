## Expected
- Exit code is 0.
- Output contains `nothing to move back` (the no-op path fires before the dry-run check).
- There is NO `dry-run: would` message — the "nothing to move back" path returns before reaching the dry-run gate because there's only one location.

## Exit Code
- 0

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "nothing to move back")
	assertNotContains(t, resp.Output, "dry-run:")
}
```
