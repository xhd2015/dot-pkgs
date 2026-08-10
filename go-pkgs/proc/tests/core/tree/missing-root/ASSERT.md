## Expected

- `resp.Procs` is empty (nil or zero-length).
- No error: missing root is soft empty, not hard fail.

## Errors

- Non-nil `err` is failure.
- Non-empty procs is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("missing root must not error: %v", err)
	}
	if len(resp.Procs) != 0 {
		t.Fatalf("missing root: got %v want empty", resp.Procs)
	}
}
```
