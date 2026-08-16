## Expected

- `err != nil`.
- RunLogin order: `bash`, then `zsh`.

## Errors

- Non-nil (last error or bash error when both fail).

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
	assertRunLoginOrder(t, resp, "bash", "zsh")
}
```
