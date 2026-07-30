## Expected

- `resp.Procs` equals the inject fixture (order and fields preserved).

## Errors

- `err` is nil.
- Live `ps` contamination (unexpected extra rows) is failure.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	assertProcsEqual(t, resp.Procs, req.ListInject)
}
```
