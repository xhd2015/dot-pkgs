## Expected

- `resp` is nil.
- Error message contains `at least one owner`.

## Side Effects

- Mock `gh` never called (`gh.called` absent).

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
		t.Fatal("expected error for empty owners")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "at least one owner") {
		t.Fatalf("expected at least one owner error, got %v", err)
	}
	if ghWasCalled(req.GhBin) {
		t.Fatal("gh should not have been called for empty owners")
	}
}```