## Expected

- `PinPatch("go1.99")` is `go1.99` (unchanged).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Pin != "go1.99" {
		t.Fatalf("PinPatch(%q) = %q, want go1.99", req.GoVersion, resp.Pin)
	}
}
```
