## Expected

- `resp` is nil.
- Error mentions owner `failuser` and stderr text `rate limit exceeded`.

## Side Effects

- Mock `gh` invoked and exited 1.

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
		t.Fatal("expected gh exit error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "failuser") {
		t.Fatalf("expected owner failuser in error, got %v", err)
	}
	if !strings.Contains(msg, "rate limit exceeded") {
		t.Fatalf("expected stderr snippet in error, got %v", err)
	}
}```