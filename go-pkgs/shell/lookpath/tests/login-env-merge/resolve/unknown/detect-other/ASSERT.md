## Expected

- `err == nil`.
- `Shell == "zsh"` (first nonempty dump in cascade).
- `Envs` contains `FOO=from-other`.
- RunLogin order: `bash`, then `zsh` (fish is not dumped directly).

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
	assertEnvsContain(t, resp.Envs, "FOO=from-other")
	assertRunLoginOrder(t, resp, "bash", "zsh")
}
```
