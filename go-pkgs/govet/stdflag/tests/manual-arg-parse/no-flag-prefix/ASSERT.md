## Expected
- No violations. Bare strings like `"daemon"` should not trigger the manual-parsing detection.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(resp.Violations) != 0 {
		t.Fatalf("expected 0 violations, got %d: %+v", len(resp.Violations), resp.Violations)
	}
}
```
