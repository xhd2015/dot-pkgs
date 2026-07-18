## Expected

1. `AttachErr` is **empty** (Attach Wait returned nil after exit marker alone).
2. Must not hang: no `timeout_hang` substring.

## Errors

- `timeout_hang: ...` — client still waits for close instead of ending on marker.
- Any other non-empty `AttachErr`.

## Side Effects

None (mock server only).

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
		t.Fatalf("Attach Wait after exit marker without close: %q (want nil)", resp.AttachErr)
	}
	if strings.Contains(strings.ToLower(resp.AttachErr), "timeout") {
		t.Fatalf("hang: %q", resp.AttachErr)
	}
}
```
