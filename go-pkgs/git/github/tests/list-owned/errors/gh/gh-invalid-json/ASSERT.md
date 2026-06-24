## Expected

- `resp` is nil.
- Error indicates JSON decode failure (`json` or `decode` in message).

## Side Effects

- Mock `gh` invoked.

## Errors

- `err` is non-nil.

## Exit Code

- N/A (library call).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected JSON decode error")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "json") && !strings.Contains(msg, "decode") {
		t.Fatalf("expected decode/json error, got %v", err)
	}
}```