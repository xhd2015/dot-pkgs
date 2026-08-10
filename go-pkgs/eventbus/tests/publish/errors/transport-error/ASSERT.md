## Expected

- `Publish` returns a non-nil error (connection refused / transport failure).
- Zero successful captures on the closed server (request never completed).

## Side Effects

- No successful HTTP exchange.

## Errors

- `err` is non-nil.

## Exit Code

- Failure from Publish (asserted as expected error).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Publish: want transport error after closed server, got nil")
	}
	// Mock was closed before Publish; capture should remain empty.
	if n := req.Capture.Len(); n != 0 {
		t.Fatalf("expected 0 captured requests on closed server, got %d", n)
	}
}
```
