## Expected

- Apply returns a path-safety error (invalid / `..` / escape).
- Stub `"not implemented"` alone must **not** pass (keeps leaf RED until real validation).

## Errors

- Non-nil error describing unsafe relative path.

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
	assertPathSafetyError(t, err, "..", "invalid", "escape")
}
```
