## Expected

- `err == nil`.
- `Path == "/tmp/from-zsh"`.
- RunLogin order: `bash`, then `zsh`.
- LookPath and RunGoEnv not called.

## Errors

- None (bash empty is soft continue).

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
	assertPath(t, resp, "/tmp/from-zsh")
	assertRunLoginOrder(t, resp, "bash", "zsh")
	assertNoLookPath(t, resp)
	assertNoGoEnv(t, resp)
}
```
