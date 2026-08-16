## Expected

- `Merged` is exactly `["FOO=1", "BAR=2"]` (nil slice ignored).

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
	assertMergedEqual(t, resp.Merged, []string{"FOO=1", "BAR=2"})
}
```
