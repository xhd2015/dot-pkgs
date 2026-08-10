## Expected

- `err == nil` (NeedsUpdate is pure bool; no error return).
- `resp.NeedsUpdate == false` when either side is empty/unparseable.

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
	assertEqual(t, "NeedsUpdate", resp.NeedsUpdate, false)
}
```
