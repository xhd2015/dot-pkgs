## Expected

- `resp.Out` is `""`.

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
	if resp.Out != "" {
		t.Fatalf("Up(0)=%q want empty", resp.Out)
	}
}
```
