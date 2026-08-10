## Expected

- `resp.Paths` equals `["/tmp/a", "/tmp/b"]` (as-is from inject; no re-parse).

## Errors

- `err` is nil.
- Empty or reordered paths is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	assertStringsEqual(t, resp.Paths, req.OpenFilesInject)
}
```
