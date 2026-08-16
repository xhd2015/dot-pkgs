## Expected

- `err != nil`.
- RunLogin called once with `bash` only (detected path does not cascade).

## Errors

- Non-nil error from shell/run failure.

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
	assertRunLoginOrder(t, resp, "bash")
}
```
