## Expected

- `err == nil`.
- `Shell == "bash"`.
- `Envs` contains `FOO=1` and `GOPATH=/tmp/gp`.
- RunLogin called once with `bash` only (no zsh cascade).
- DetectShell called at least once.

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
	assertShell(t, resp, "bash")
	assertEnvsContain(t, resp.Envs, "FOO=1", "GOPATH=/tmp/gp")
	assertRunLoginOrder(t, resp, "bash")
	if resp.DetectShellCalls < 1 {
		t.Fatalf("DetectShellCalls = %d, want >= 1", resp.DetectShellCalls)
	}
}
```
