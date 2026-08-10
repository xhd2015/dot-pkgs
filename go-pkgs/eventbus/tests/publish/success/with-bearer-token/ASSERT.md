## Expected

- `Publish` returns nil.
- Captured request has `Authorization: Bearer secret-token`.
- Path remains `/publish` and method `POST`.

## Side Effects

- One authenticated POST to the mock hub.

## Errors

- `err` is nil.

## Exit Code

- Success.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Publish with token: unexpected error: %v", err)
	}
	if req.Capture.Len() != 1 {
		t.Fatalf("expected exactly 1 HTTP request, got %d", req.Capture.Len())
	}
	got, ok := req.Capture.Last()
	if !ok {
		t.Fatal("missing captured request")
	}
	want := "Bearer secret-token"
	if got.Authorization != want {
		t.Fatalf("Authorization: got %q, want %q", got.Authorization, want)
	}
	if got.Method != "POST" || got.Path != "/publish" {
		t.Fatalf("request %s %s, want POST /publish", got.Method, got.Path)
	}
}
```
