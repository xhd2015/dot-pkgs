## Expected
- Two violations from the `err-unchecked-or-ignored` checker.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(resp.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %+v", len(resp.Violations), resp.Violations)
	}
	for i, v := range resp.Violations {
		if v.Checker != "err-unchecked-or-ignored" {
			t.Fatalf("violation %d: expected checker 'err-unchecked-or-ignored', got %q", i, v.Checker)
		}
	}
}
```
