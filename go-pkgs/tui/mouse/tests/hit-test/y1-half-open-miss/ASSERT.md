## Expected

- `err` is nil.
- `resp.HitOK` is false (half-open: y < y1 required).

## Errors

- Treating y1 as inclusive would falsely hit the run chip.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("HitTest: %v", err)
	}
	if resp.HitOK {
		t.Fatalf("HitTest(%d,%d) hit id=%q, want miss at half-open y1",
			req.X, req.LocalY, resp.Hit.ID)
	}
}
```
