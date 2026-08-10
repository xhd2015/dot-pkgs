## Expected

- `err == nil`.
- One found item; Path under ExtraDirs; `From == ""`.
- Item invariants hold.

## Errors

- None.

```go
import (
	"path/filepath"
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
	want := filepath.Join(req.ExtraDirs[0], "mytool")
	assertItemFound(t, resp.Items[0], "mytool", want, "")
}
```
