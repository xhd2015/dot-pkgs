## Expected

- `SetTitle` returns a non-nil error for empty title (no successful set).

## Errors

- Empty title validation error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for empty title")
	}
}
```
