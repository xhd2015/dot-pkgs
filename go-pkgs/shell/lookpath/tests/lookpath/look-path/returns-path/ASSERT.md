## Expected

- `err == nil`.
- `resp.Path` equals the injected LookPath hit.
- Via is not required on this API (LookPath returns string only).

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
	assertPathEqual(t, resp.Path, req.LookPathHit)
}
```
