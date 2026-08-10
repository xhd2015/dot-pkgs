## Expected

- `err == nil`.
- `resp.Path` equals the absolute executable path.
- `resp.Via == "direct"`.
- `resp.LookPathCalls` is empty (no PATH fallthrough).

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
	assertPathEqual(t, resp.Path, req.Name)
	assertEqual(t, "Via", resp.Via, "direct")
	assertNoLookPathCalls(t, resp)
}
```
