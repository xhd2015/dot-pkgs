## Expected

- `err != nil`.
- `resp.LookPathCalls` is empty (no fallthrough after direct-path miss).

## Errors

- Non-nil error for non-executable absolute path.

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
