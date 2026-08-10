## Expected

- `err == nil`.
- Missing item.
- `DirsEnv == ""`.

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
	_ = req
	assertNoError(t, err)
	assertItemInvariants(t, resp.Items)
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}
	assertItemMissing(t, resp.Items[0], "ghost")
	assertEqual(t, "DirsEnv", resp.DirsEnv, "")
}
```
