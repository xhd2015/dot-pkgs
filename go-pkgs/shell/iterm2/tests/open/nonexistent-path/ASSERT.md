## Expected

- Non-nil error from stat; osascript not invoked.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected stat error")
	}
	if resp.CapturedScript != "" {
		t.Fatal("osascript should not run")
	}
}
```