## Expected

- `err` is nil.
- `resp.Index` is **1** from first display’s [3, 132, 234] only.
- Second display’s type-0 ids must not renumber first-display indices (not 3 or higher).

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	if err != nil {
		t.Fatalf("SpaceIndexForWindow: unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.Index != 1 {
		t.Fatalf("Index=%d want 1 (first display only; second display must not shift indices)", resp.Index)
	}
}
```
