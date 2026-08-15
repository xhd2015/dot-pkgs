## Expected

- Uses `tell aSession` and `on error`; must not use invalid `of aSession` access.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !scriptUsesTellSession(resp.Script) {
		t.Fatalf("expected tell aSession access: %q", resp.Script)
	}
	if !scriptUsesOnError(resp.Script) {
		t.Fatalf("expected on error handler: %q", resp.Script)
	}
}
```