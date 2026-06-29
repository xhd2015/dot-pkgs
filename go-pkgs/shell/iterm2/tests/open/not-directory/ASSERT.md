## Expected

- Error mentions not a directory; no script captured.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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