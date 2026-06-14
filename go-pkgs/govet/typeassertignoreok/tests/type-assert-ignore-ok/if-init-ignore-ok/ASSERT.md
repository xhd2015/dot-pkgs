## Expected
- One violation from the `type-assert-ignore-ok` checker.

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
	if v.Checker != "type-assert-ignore-ok" {
		t.Fatalf("expected checker 'type-assert-ignore-ok', got %q", v.Checker)
	}
}
```
