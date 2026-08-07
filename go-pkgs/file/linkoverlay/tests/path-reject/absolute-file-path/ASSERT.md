## Expected

- Apply returns a path-safety error (absolute / invalid).
- Stub `"not implemented"` alone must **not** pass.

## Errors

- Non-nil error describing absolute overlay path.

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
	assertPathSafetyError(t, err, "absolute", "invalid")
}
```
