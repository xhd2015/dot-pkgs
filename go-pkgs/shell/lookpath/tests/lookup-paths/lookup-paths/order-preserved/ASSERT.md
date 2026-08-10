## Expected

- `err == nil`.
- `len(Items) == 2`.
- `Items[0].Name == "beta"`, `Items[1].Name == "alpha"` (input order, not sorted).
- Both found with paths matching LookPathHits; `From == ""`.
- Item invariants hold.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertItemInvariants(t, resp.Items)
	if len(resp.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(resp.Items))
	}
	assertItemFound(t, resp.Items[0], "beta", req.LookPathHits["beta"], "")
	assertItemFound(t, resp.Items[1], "alpha", req.LookPathHits["alpha"], "")
}
```
