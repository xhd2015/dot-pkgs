## Expected

- `err == nil`.
- `resp.DestSize > 0`.
- Dest path exists on disk.

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
	if resp.DestSize <= 0 {
		t.Fatalf("DestSize = %d, want > 0", resp.DestSize)
	}
	assertFileExists(t, resp.ZipPath)
}
```
