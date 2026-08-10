## Expected

- `err != nil`.
- Empty or missing `X.Y.Z` is a detectable error (not silent success).

## Errors

- Non-nil error from unparseable empty input.

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
