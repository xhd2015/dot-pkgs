## Expected

- `Publish` returns a non-nil error (non-2xx is not success).
- The mock still received the POST (request was attempted).

## Side Effects

- One POST to the mock hub that returned 500.

## Errors

- `err` is non-nil.

## Exit Code

- Failure from Publish (asserted as expected error).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Publish: want error on non-2xx status, got nil")
	}
	if req.Capture.Len() != 1 {
		t.Fatalf("expected 1 captured request before error, got %d", req.Capture.Len())
	}
	got, _ := req.Capture.Last()
	if got.Method != "POST" || got.Path != "/publish" {
		t.Fatalf("request %s %s, want POST /publish", got.Method, got.Path)
	}
}
```
