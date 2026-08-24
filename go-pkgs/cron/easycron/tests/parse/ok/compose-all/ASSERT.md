## Expected

- err is nil
- Align 0, Until 19:00, Quiet set

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
	if resp.Expr.Interval != 5*time.Minute {
		t.Fatalf("Interval=%v", resp.Expr.Interval)
	}
	if resp.Expr.Align == nil || *resp.Expr.Align != 0 {
		t.Fatalf("Align=%v, want 0", resp.Expr.Align)
	}
	if resp.Expr.Until == nil || resp.Expr.Until.Hour != 19 {
		t.Fatalf("Until=%v", resp.Expr.Until)
	}
	if resp.Expr.Quiet == nil {
		t.Fatal("Quiet nil")
	}
}
```
