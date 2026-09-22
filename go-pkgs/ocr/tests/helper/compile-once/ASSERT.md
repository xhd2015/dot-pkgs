## Expected

- `err` is nil.
- Both recognizes return `fake-vision`.
- Fake `swiftc` counter is `1`.

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
	if resp.Text != "fake-vision" || resp.Text2 != "fake-vision" {
		t.Fatalf("texts = %q / %q", resp.Text, resp.Text2)
	}
	if resp.SwiftcN != "1" {
		t.Fatalf("swiftc count = %q, want 1", resp.SwiftcN)
	}
	if resp.Engine != "vision" {
		t.Fatalf("Engine = %q", resp.Engine)
	}
}
```
