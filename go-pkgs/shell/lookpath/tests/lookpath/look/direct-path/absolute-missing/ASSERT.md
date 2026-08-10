## Expected

- `err != nil`.
- Error message includes the requested path or basename.
- `resp.LookPathCalls` is empty.

## Errors

- Non-nil resolution error; no later-stage success.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	assertError(t, err)
	base := filepath.Base(req.Name)
	msg := err.Error()
	if !strings.Contains(msg, req.Name) && !strings.Contains(msg, base) {
		t.Fatalf("error %q should mention path %q or base %q", msg, req.Name, base)
	}
	assertNoLookPathCalls(t, resp)
}
```
