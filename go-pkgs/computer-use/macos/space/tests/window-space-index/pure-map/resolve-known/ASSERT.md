## Expected

- `err` is nil.
- `resp.Index` is **1** (middle of [3, 132, 234]).

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
		t.Fatalf("ResolveWindowSpaceIndex: unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.Index != 1 {
		t.Fatalf("Index=%d want 1", resp.Index)
	}
}
```
