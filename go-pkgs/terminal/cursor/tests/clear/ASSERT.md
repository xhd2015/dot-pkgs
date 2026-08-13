## Expected

- `resp.Out` is `"\r\x1b[2K"`.

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
	const want = "\r\x1b[2K"
	if resp.Out != want {
		t.Fatalf("Clear()=%q want %q", resp.Out, want)
	}
}
```
