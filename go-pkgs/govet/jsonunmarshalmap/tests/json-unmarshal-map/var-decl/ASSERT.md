## Expected
- One violation from the `json-unmarshal-map` checker.
- The violation message mentions using a typed struct instead.

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
	if v.Checker != "json-unmarshal-map" {
		t.Fatalf("expected checker 'json-unmarshal-map', got %q", v.Checker)
	}
}
```
