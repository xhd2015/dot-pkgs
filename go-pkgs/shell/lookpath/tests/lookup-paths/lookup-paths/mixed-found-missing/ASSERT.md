## Expected

- `err == nil` (best-effort: missing names are not errors).
- Two items in input order: found then missing.
- Missing item has empty Path and From.
- Item invariants hold.

## Errors

- None (missing is represented as Missing=true, not err).

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
	assertItemFound(t, resp.Items[0], "found", req.LookPathHits["found"], "")
	assertItemMissing(t, resp.Items[1], "ghost")
}
```
