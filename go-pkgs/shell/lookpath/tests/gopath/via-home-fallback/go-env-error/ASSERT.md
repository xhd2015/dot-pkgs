## Expected

- `err == nil`.
- `Path == filepath.Join(Home, "go")`.
- LookPath and RunGoEnv were called; go-env failure does not surface.

## Errors

- None at ResolveGoPathWith (RunGoEnv error is soft).

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
	if len(resp.GoEnvBins) == 0 {
		t.Fatal("RunGoEnv was not called")
	}
	assertEqual(t, "GoEnvBins[0]", resp.GoEnvBins[0], req.GoBin)
}
```
