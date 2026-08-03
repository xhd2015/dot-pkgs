## Expected

- `err` is nil.
- `resp.Index` is **1** (dense over type-0 only: 3→0, 132→1, 234→2).
- Must **not** be 2 (would mean type-4 was counted).

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
		t.Fatalf("Index=%d want 1 (type!=0 must be skipped for dense indexing)", resp.Index)
	}
}
```
