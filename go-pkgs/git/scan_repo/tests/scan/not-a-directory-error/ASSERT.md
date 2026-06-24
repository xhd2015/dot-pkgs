## Expected

- `resp` is nil.

## Errors

- `err` is non-nil.
- Error message indicates not a directory.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for file root")
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "directory") {
		t.Fatalf("error should mention directory, got: %v", err)
	}
}
```