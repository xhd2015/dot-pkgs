## Expected

- Error mentions osascript; script was built and passed to hook.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected osascript error")
	}
	if !strings.Contains(err.Error(), "osascript") {
		t.Fatalf("error = %v", err)
	}
	if resp.CapturedScript == "" {
		t.Fatal("expected script to reach osascript hook")
	}
}
```