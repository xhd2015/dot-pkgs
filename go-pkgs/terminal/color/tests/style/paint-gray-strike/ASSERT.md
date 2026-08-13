## Expected

- `err` is nil.
- `resp.Out` equals exactly `"\x1b[90m\x1b[9mhello\x1b[0m"` (one reset).

## Side Effects

- None.

## Errors

- None.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Paint(Gray, Strike): unexpected error: %v", err)
	}
	const want = "\x1b[90m\x1b[9mhello\x1b[0m"
	if resp.Out != want {
		t.Fatalf("Paint(Gray, Strike): got %q, want %q", resp.Out, want)
	}
}
```
