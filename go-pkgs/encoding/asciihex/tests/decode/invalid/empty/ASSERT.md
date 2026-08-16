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
		t.Fatal("Decode(\"\"): want error, got nil")
	}
	const want = "invalid hex escape sequence"
	if got := err.Error(); got != want {
		t.Fatalf("Decode(\"\"): error %q, want %q", got, want)
	}
}
```
