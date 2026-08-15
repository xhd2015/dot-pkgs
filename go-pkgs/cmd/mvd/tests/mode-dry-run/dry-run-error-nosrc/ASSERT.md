## Expected
- Non-zero exit code (validation fails before dry-run check).
- Output contains `does not exist`.
- Output does NOT contain `dry-run:` (dry-run message only prints after validation passes).

## Exit Code
- Non-zero

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0: %s", resp.Output)
	}
	assertContains(t, resp.Output, "does not exist")
	assertNotContains(t, resp.Output, "dry-run:")
}
```
