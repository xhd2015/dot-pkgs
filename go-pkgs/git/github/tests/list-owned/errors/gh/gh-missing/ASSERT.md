## Expected

- `resp` is nil.
- Error contains `gh not found` (case-insensitive match).

## Side Effects

- No successful gh execution.

## Errors

- `err` is non-nil.

## Exit Code

- N/A (library call).

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for missing gh binary")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "gh not found") {
		t.Fatalf("expected gh not found error, got %v", err)
	}
}```