## Expected

- `err != nil`
- error mentions `HTTP 500`
- single GET (no auth retry)

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
	assertContains(t, "err", err.Error(), "HTTP 500")
	assertEqual(t, "GetCount", resp.GetCount, 1)
}
```
