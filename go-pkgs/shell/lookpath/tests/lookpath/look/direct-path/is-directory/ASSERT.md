## Expected

- `err != nil`.
- `resp.LookPathCalls` is empty.

## Errors

- Non-nil error when the direct path is a directory.

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
	assertNoLookPathCalls(t, resp)
}
```
