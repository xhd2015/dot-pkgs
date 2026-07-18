## Expected

1. `AttachErr` is **empty** (Attach Wait returned nil).
2. Session may remain listed with `status=exited` after attach returns.

## Errors

- `AttachErr` contains `terminal closed` and/or `unexpected EOF` — current bug
  when server closes without close frame 1000.
- Any non-nil attach error after a normal shell exit.

## Side Effects

- Test cleanup deletes the session via REST.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp == nil {
		t.Fatal("nil response")
	}
	if resp.AttachErr != "" {
		t.Fatalf("Attach Wait error after shell exit: %q (want nil; often terminal closed / unexpected EOF when CloseCode is 1006)",
			resp.AttachErr)
	}
	// Extra guard: even if formatting changes, these substrings must not appear.
	lower := strings.ToLower(resp.AttachErr)
	if strings.Contains(lower, "unexpected eof") || strings.Contains(lower, "terminal closed") {
		t.Fatalf("AttachErr still looks like unclean close: %q", resp.AttachErr)
	}
}
```
