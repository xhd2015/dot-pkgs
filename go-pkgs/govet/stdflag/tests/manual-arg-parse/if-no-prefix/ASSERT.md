## Expected
- No violations. `if arg == "status"` without `-` prefix should not trigger detection.

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
