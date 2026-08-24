## Expected

- err is nil
- Interval 1h, Align 4m

## Side Effects

- None.

## Errors

- None unless noted in Expected.

## Exit Code

- N/A (in-process library).

```go
import (
	"testing"
	"time"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Parse(%q): %v", req.Expr, err)
	}
	if resp.Expr.Interval != time.Hour {
		t.Fatalf("Interval=%v, want 1h", resp.Expr.Interval)
	}
	if resp.Expr.Align == nil || *resp.Expr.Align != 4*time.Minute {
		t.Fatalf("Align=%v, want 4m", resp.Expr.Align)
	}
}
```
