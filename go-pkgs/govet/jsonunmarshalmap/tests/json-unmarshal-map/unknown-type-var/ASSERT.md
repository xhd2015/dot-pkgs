## Expected
- No violations are reported (type is unknown, checker should not false-positive).

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(resp.Violations) != 0 {
		t.Fatalf("expected 0 violations, got %d: %+v", len(resp.Violations), resp.Violations)
	}
}
```
