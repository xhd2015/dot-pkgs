## Expected

- `err` is nil.
- `resp.Out` equals exactly `"\x1b[32mhello\x1b[0m"`.

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
		t.Fatalf("Style{Enabled:true}.Green(%q): unexpected error: %v", req.Text, err)
	}
	const want = "\x1b[32mhello\x1b[0m"
	if resp.Out != want {
		t.Fatalf("Style{Enabled:true}.Green(%q): got %q, want %q", req.Text, resp.Out, want)
	}
}
```
