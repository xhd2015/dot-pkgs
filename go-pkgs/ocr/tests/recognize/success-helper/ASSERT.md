## Expected

- `err` is nil.
- `resp.Text` is `hello vision`.
- `resp.Engine` is `vision`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Text != "hello vision" {
		t.Fatalf("Text = %q", resp.Text)
	}
	if resp.Engine != "vision" {
		t.Fatalf("Engine = %q", resp.Engine)
	}
}
```
