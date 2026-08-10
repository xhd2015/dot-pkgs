## Expected

- CopySpaces hook was invoked (`CopySpacesCalled`).
- `CapturedMask` is **7** (locked experiment mask for `CGSCopySpacesForWindows`).
- Resolve still succeeds: `Index == 1` for space id 132.

## Errors

- None for this happy path; mask contract is the primary assert.

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
	if !resp.CopySpacesCalled {
		t.Fatal("expected WithCopySpacesForWindows spy to be called")
	}
	if resp.CapturedMask != 7 {
		t.Fatalf("CapturedMask=%d want 7", resp.CapturedMask)
	}
	if resp.Index != 1 {
		t.Fatalf("Index=%d want 1 (spy returned space 132)", resp.Index)
	}
}
```
