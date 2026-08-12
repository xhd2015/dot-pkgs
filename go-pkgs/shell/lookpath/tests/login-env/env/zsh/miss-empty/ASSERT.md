## Expected

- `err == nil` (miss is cascade-friendly, not an error).
- `Value == ""`.

## Errors

- None (unset/empty is not an error).

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
	assertEqual(t, "Value", resp.Value, "")
	assertRunLoginShell(t, resp, "zsh")
}
```
