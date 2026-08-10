## Expected

- `err == nil`.
- `len(resp.Items) == 0`.

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
	if len(resp.Items) != 0 {
		t.Fatalf("Items = %#v, want empty", resp.Items)
	}
}
```
