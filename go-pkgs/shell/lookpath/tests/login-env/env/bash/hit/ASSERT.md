## Expected

- `err == nil`.
- `Value == "1"`.
- RunLogin first shell is `bash`.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	assertEqual(t, "Value", resp.Value, "1")
	assertRunLoginShell(t, resp, "bash")
}
```
