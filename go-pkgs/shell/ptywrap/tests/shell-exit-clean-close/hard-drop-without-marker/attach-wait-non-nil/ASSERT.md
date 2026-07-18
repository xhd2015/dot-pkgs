## Expected

1. `AttachErr` is **non-empty** (Attach Wait did not treat hard drop as success).
2. Must not hang: no `timeout_hang` substring alone as the only outcome without
   an error path — if hang occurs, fail.

## Errors

- Empty `AttachErr` — client incorrectly treats bare drop as clean exit.
- `timeout_hang: ...` — Attach never returned (also a failure for this leaf).

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
	if resp.AttachErr == "" {
		t.Fatal("Attach Wait after hard drop without marker: empty AttachErr (want non-nil error; do not silence 1006/EOF without exit marker)")
	}
	if strings.HasPrefix(resp.AttachErr, "timeout_hang:") {
		t.Fatalf("hang waiting for attach after hard drop: %q", resp.AttachErr)
	}
}
```
