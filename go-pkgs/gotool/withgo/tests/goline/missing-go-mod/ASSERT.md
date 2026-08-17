## Expected

- `ModuleGoLine` returns an error (missing go.mod).
- `GoLine` is empty.

## Errors

- `resp.Err` is non-nil. Harness `err` is nil.

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
	if resp.Err == nil {
		t.Fatalf("ModuleGoLine(%q) succeeded with %q, want error (missing go.mod)", req.ModDir, resp.GoLine)
	}
	if resp.GoLine != "" {
		t.Fatalf("ModuleGoLine(%q) GoLine = %q, want empty on error", req.ModDir, resp.GoLine)
	}
}
```
