## Expected

- `err == nil`.
- One item: found, Path = LookPath hit, `From == ""`.
- LookPath was called with `"mytool"`.
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
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}
	assertItemFound(t, resp.Items[0], "mytool", req.LookPathHits["mytool"], "")
	if len(resp.LookPathCalls) == 0 {
		t.Fatal("expected LookPath to be called")
	}
	assertEqual(t, "LookPathCalls[0]", resp.LookPathCalls[0], "mytool")
}
```
