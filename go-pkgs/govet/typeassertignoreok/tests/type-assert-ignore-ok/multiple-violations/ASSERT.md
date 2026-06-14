## Expected
- Two violations from the `type-assert-ignore-ok` checker.

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
		if v.Checker != "type-assert-ignore-ok" {
			t.Fatalf("expected checker 'type-assert-ignore-ok', got %q", v.Checker)
		}
	}
}
```
