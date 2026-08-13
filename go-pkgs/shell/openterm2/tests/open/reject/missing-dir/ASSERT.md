## Expected

- `err != nil`.
- Neither `OpenITerm` nor `OpenTerminal` is called.

## Side Effects

- No opener hook runs.

## Errors

- Non-nil validation error for a missing path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertError(t, err)
	assertNoOpeners(t, resp)
}
```
