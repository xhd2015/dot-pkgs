## Expected
- Two violations (one per independent for/switch or for/if block).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(resp.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %+v", len(resp.Violations), resp.Violations)
	}
	for _, v := range resp.Violations {
		if v.Checker != "manual-flag-parse" {
			t.Fatalf("expected all violations to be 'manual-flag-parse', got %q", v.Checker)
		}
	}
}
```
