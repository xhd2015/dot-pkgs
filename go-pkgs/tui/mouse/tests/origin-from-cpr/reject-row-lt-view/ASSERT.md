## Expected

- `err` is nil.
- `resp.OriginOK` is false (live reject; not OriginFromCPRClamped).

## Errors

- Accepting row1 < viewLines as origin 0 would mark a stale probe as top-anchored.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("OriginFromCPR: %v", err)
	}
	if resp.OriginOK {
		t.Fatalf("OriginFromCPR(%d,%d) ok with origin %d; live rule must reject row1 < viewLines",
			req.Row1, req.ViewLines, resp.OriginY0)
	}
}
```
