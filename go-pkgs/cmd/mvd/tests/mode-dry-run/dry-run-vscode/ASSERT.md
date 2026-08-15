## Expected
- Exit code is 0.
- Output contains `dry-run: would open VSCode`.
- No `code` process is actually launched.

## Exit Code
- 0

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertContains(t, resp.Output, "dry-run: would open VSCode at")
}
```
