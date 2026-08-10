## Expected

- `err != nil`.

## Errors

- Non-nil error from failed HTTP resolve.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	assertError(t, err)
}
```
