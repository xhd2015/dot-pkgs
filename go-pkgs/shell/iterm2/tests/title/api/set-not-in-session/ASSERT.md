## Expected

- `SetTitle` returns a non-nil error when `ITERM_SESSION_ID` is empty.

## Errors

- Not-in-session error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error when ITERM_SESSION_ID is empty")
	}
}
```
