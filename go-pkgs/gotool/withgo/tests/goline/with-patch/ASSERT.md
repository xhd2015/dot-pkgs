## Expected

- `ModuleGoLine` returns `go1.19` (patch dropped).

## Errors

- `err` and `resp.Err` are nil.

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
	if resp.Err != nil {
		t.Fatalf("ModuleGoLine(%q) failed: %v", req.ModDir, resp.Err)
	}
	if resp.GoLine != "go1.19" {
		t.Fatalf("ModuleGoLine(%q) = %q, want go1.19", req.ModDir, resp.GoLine)
	}
}
```
