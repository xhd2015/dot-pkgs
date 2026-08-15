## Expected
- One violation from the `err-unchecked-or-ignored` checker.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(resp.Violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %+v", len(resp.Violations), resp.Violations)
	}
	v := resp.Violations[0]
	if v.Checker != "err-unchecked-or-ignored" {
		t.Fatalf("expected checker 'err-unchecked-or-ignored', got %q", v.Checker)
	}
}
```
