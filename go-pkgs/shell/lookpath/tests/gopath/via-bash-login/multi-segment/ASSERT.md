## Expected

- `err == nil`.
- `Path == "/tmp/a"` (first colon segment only; not `/tmp/b`).
- RunLogin only `bash`; no LookPath / RunGoEnv.

## Errors

- None.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	_ = req
	assertNoError(t, err)
	// Spec: multi-GOPATH a:b → filepath.Clean of first segment.
	want := filepath.Clean("/tmp/a")
	assertPath(t, resp, want)
	assertRunLoginOrder(t, resp, "bash")
	assertNoLookPath(t, resp)
	assertNoGoEnv(t, resp)
}
```
