## Expected

- err is nil
- Quiet Start 19:00 End 06:30

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
	q := resp.Expr.Quiet
	if q == nil {
		t.Fatal("Quiet is nil")
	}
	if q.Start.Hour != 19 || q.Start.Minute != 0 || q.End.Hour != 6 || q.End.Minute != 30 {
		t.Fatalf("Quiet=%+v, want 19:00→06:30", q)
	}
}
```
