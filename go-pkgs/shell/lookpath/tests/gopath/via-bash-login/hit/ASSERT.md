## Expected

- `err == nil`.
- `Path == "/tmp/from-bash"`.
- RunLogin called only for `bash` (cascade short-circuit before zsh).
- LookPath and RunGoEnv not called.

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
	assertPath(t, resp, "/tmp/from-bash")
	assertRunLoginOrder(t, resp, "bash")
	assertNoLookPath(t, resp)
	assertNoGoEnv(t, resp)
}
```
