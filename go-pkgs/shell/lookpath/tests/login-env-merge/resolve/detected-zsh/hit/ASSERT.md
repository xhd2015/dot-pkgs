## Expected

- `err == nil`.
- `Shell == "zsh"`.
- `Envs` contains `FOO=1` and `GOPATH=/tmp/gp`.
- RunLogin called once with `zsh` only.

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
	assertShell(t, resp, "zsh")
	assertEnvsContain(t, resp.Envs, "FOO=1", "GOPATH=/tmp/gp")
	assertRunLoginOrder(t, resp, "zsh")
}
```
