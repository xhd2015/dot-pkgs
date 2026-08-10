## Expected

- `err` is nil.
- `resp.Index` is **2** (last of three user Desktops).

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("SpaceIndexForWindow: unexpected err: %v", err)
	}
	if resp == nil {
		t.Fatal("nil Response")
	}
	if resp.Index != 2 {
		t.Fatalf("Index=%d want 2", resp.Index)
	}
}
```
