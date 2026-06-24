## Expected

- `resp` is nil.

## Errors

- `err` is non-nil.
- Error message mentions at least one root is required.

## Exit Code

- N/A (library returns error, not exit code).

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for empty roots")
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "root") {
		t.Fatalf("error should mention root, got: %v", err)
	}
}
```