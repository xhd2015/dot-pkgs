## Expected

- No error.
- Runner called once with `-n` instead of `-r`.

## Side Effects

- One runner call.

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
	assertNoError(t, err)
	assertDefaultRunner(t, resp, req, []string{req.CodePath, "-n", req.Dir})
}
```
