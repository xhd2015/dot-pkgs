## Expected

- `err == nil`.
- `Envs` contains `FOO=1` and `GOPATH=/tmp/gp`.
- RunLogin first shell is `bash`.
- RunLogin env includes `HOME=<req.Home>`.

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
	assertEnvsContain(t, resp.Envs, "FOO=1", "GOPATH=/tmp/gp")
	assertRunLoginShell(t, resp, "bash")
	assertRunLoginEnvHasHome(t, resp, req.Home)
}
```
