## Expected

- `resp` is nil.
- Error message contains `invalid owner`.

## Side Effects

- Mock `gh` never called.

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
		t.Fatal("expected error for empty owner string")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "invalid owner") {
		t.Fatalf("expected invalid owner error, got %v", err)
	}
	if ghWasCalled(req.GhBin) {
		t.Fatal("gh should not have been called for invalid owner string")
	}
}```