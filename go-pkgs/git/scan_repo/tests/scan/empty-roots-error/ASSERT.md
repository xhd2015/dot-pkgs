## Expected

- Scan returns an error; `Run` may still populate `resp.Elapsed`.

## Errors

- `err` is non-nil.
- Error message mentions at least one root is required.

## Exit Code

- N/A (library returns error, not exit code).

```go
import (
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected error for empty roots")
	}
	msg := strings.ToLower(err.Error())
	if !strings.Contains(msg, "root") {
		t.Fatalf("error should mention root, got: %v", err)
	}
}
```