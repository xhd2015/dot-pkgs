## Expected

- NextOK true, Next >= 2026-08-25T06:30:00Z and Active

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
		t.Fatal(err)
	}
	if !resp.NextOK {
		t.Fatal("expected a fire after quiet")
	}
	min := time.Date(2026, 8, 25, 6, 30, 0, 0, utc())
	if resp.Next.Before(min) {
		t.Fatalf("Next %v before resume %v", resp.Next, min)
	}
	if !resp.Expr.Active(resp.Next, utc()) {
		t.Fatalf("Next %v not active", resp.Next)
	}
}
```
