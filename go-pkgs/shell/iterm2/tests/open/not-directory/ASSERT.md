## Expected

- Error mentions not a directory; no script captured.

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for file path")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("error = %v", err)
	}
	if resp.CapturedScript != "" {
		t.Fatal("osascript should not run")
	}
}
```