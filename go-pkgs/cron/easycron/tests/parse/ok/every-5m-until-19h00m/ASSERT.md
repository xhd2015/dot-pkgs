## Expected

- err is nil
- Interval 5m, Until 19:00

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
		t.Fatalf("Interval=%v, want 5m", resp.Expr.Interval)
	}
	if resp.Expr.Until == nil || resp.Expr.Until.Hour != 19 || resp.Expr.Until.Minute != 0 {
		t.Fatalf("Until=%v, want 19:00", resp.Expr.Until)
	}
}
```
