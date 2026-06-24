## Expected

- `resp` is nil.

## Errors

- `err` is non-nil.
- Error message contains the missing root path.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for missing root")
	}
	if resp != nil {
		t.Fatalf("expected nil response, got %v", resp)
	}
	if !strings.Contains(err.Error(), req.Roots[0]) {
		t.Fatalf("error should contain root path %q, got: %v", req.Roots[0], err)
	}
}
```