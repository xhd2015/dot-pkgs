## Expected

- `resp.Out` is `"\x1b[2A"`.

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
	const want = "\x1b[2A"
	if resp.Out != want {
		t.Fatalf("Up(2)=%q want %q", resp.Out, want)
	}
}
```
