## Expected
- One violation from the `manual-flag-parse` checker, triggered by the `-v` single-dash prefix.

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
	if v.Checker != "manual-flag-parse" {
		t.Fatalf("expected checker 'manual-flag-parse', got %q", v.Checker)
	}
}
```
