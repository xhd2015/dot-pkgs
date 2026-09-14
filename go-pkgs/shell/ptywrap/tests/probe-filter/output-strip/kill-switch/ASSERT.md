## Expected

- `err` is nil.
- `resp.Out` equals `req.Data`.
- `resp.Rest` is empty.

```go
import (
	"bytes"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(resp.Out, req.Data) {
		t.Fatalf("out=%q want %q", resp.Out, req.Data)
	}
	if len(resp.Rest) != 0 {
		t.Fatalf("rest=%q", resp.Rest)
	}
}
```
