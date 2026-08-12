## Expected

- `err == nil`.
- `Path == "/tmp/from-go"` (trimmed).
- RunLogin order: `bash`, `zsh`.
- LookPath called with `"go"`; RunGoEnv called with GoBin.

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
	assertPath(t, resp, "/tmp/from-go")
	assertRunLoginOrder(t, resp, "bash", "zsh")
	assertLookPathGo(t, resp)
	if len(resp.GoEnvBins) == 0 {
		t.Fatal("RunGoEnv was not called")
	}
	assertEqual(t, "GoEnvBins[0]", resp.GoEnvBins[0], req.GoBin)
}
```
