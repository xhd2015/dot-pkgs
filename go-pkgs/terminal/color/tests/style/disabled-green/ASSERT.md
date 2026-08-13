## Expected

- `err` is nil.
- `resp.Out` equals `"hello"` with no SGR prefix or reset.

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
		t.Fatalf("Style{Enabled:false}.Green(%q): unexpected error: %v", req.Text, err)
	}
	if resp.Out != "hello" {
		t.Fatalf("Style{Enabled:false}.Green(%q): got %q, want %q (plain text)", req.Text, resp.Out, "hello")
	}
}
```
