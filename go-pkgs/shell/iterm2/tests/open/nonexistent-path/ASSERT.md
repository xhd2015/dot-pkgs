## Expected

- Non-nil error from stat; osascript not invoked.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected stat error")
	}
	if resp.CapturedScript != "" {
		t.Fatal("osascript should not run")
	}
}
```