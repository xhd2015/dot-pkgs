## Expected
- mvd reports that the basename is ambiguous since two tracked projects share it.
- The output contains "ambiguous root basename kool".

## Exit Code
- Non-zero

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0: %s", resp.Output)
	}
	assertContains(t, resp.Output, "ambiguous root basename kool")
}
```
