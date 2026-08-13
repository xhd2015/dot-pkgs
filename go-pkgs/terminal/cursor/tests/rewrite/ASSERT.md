## Expected

- `resp.Out` is `"\r\x1b[2Khi"`.

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
	const want = "\r\x1b[2Khi"
	if resp.Out != want {
		t.Fatalf("Rewrite(%q)=%q want %q", req.Text, resp.Out, want)
	}
}
```
