## Expected
- One violation (one per for/switch pattern, not one per case).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
