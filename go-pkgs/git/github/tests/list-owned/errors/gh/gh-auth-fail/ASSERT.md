## Expected

- `resp` is nil.
- Error contains `gh auth login` hint.

## Side Effects

- Mock `gh` invoked and exited 4.

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
		t.Fatal("expected auth failure error")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "gh auth login") {
		t.Fatalf("expected gh auth login hint, got %v", err)
	}
}```