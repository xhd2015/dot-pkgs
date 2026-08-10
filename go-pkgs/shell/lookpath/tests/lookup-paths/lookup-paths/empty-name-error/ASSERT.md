## Expected

- `err != nil` (empty name is invalid).
- Partial success not required; error is enough.

## Errors

- Non-nil error for empty name element.

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
