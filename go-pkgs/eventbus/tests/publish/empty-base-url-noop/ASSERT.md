## Expected

- `Publish` returns nil (no-op success).
- `req.Capture` has zero recorded requests (no HTTP performed).

## Side Effects

- No network I/O to a publish endpoint.

## Errors

- `err` is nil.

## Exit Code

- Success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("empty base URL Publish: want nil error, got %v", err)
	}
	if n := req.Capture.Len(); n != 0 {
		t.Fatalf("empty base URL must not perform HTTP; captured %d request(s)", n)
	}
}
```
