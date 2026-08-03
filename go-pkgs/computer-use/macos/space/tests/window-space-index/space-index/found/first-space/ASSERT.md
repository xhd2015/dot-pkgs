## Expected

- `err` is nil.
- `resp.Index` is **0**.

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
	if resp.Index != 0 {
		t.Fatalf("Index=%d want 0", resp.Index)
	}
}
```
