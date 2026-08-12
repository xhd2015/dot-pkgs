## Expected

- `err == nil`.
- `Path == filepath.Join(Home, "go")`.
- LookPath was attempted for `"go"`.
- RunGoEnv not required (miss short-circuits go env).

## Errors

- None (LookPath miss is soft).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertNoError(t, err)
	assertPath(t, resp, homeGo(req))
	assertRunLoginOrder(t, resp, "bash", "zsh")
	assertLookPathGo(t, resp)
	assertNoGoEnv(t, resp)
}
```
