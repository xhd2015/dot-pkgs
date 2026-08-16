## Expected

- `err` is non-nil.
- `err.Error()` equals exactly `invalid hex escape sequence`.

## Side Effects

- None.

## Errors

- Exact string: `invalid hex escape sequence`.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("Decode(%q): want error, got nil", req.Hex)
	}
	const want = "invalid hex escape sequence"
	if got := err.Error(); got != want {
		t.Fatalf("Decode(%q): error %q, want %q", req.Hex, got, want)
	}
}
```
