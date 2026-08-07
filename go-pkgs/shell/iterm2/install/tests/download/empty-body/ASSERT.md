## Expected

- `err != nil` (empty body is not a valid zip download).

## Errors

- Non-nil download error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	_ = resp
	assertError(t, err)
}
```
