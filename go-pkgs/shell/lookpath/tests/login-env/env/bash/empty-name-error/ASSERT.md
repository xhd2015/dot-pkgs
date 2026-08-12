## Expected

- `err != nil` (empty name is invalid).
- Value not required when error.

## Errors

- Non-nil error for empty env name.

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
