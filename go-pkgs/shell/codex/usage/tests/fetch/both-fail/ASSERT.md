## Expected

- `err != nil`
- error mentions fetch / HTTP

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
	assertContains(t, "err", err.Error(), "fetch")
}
```
