## Expected
- No violations. A switch outside a for loop, even with `--` prefix, is not manual flag parsing.

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
