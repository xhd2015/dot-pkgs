## Expected
- Two violations from the `json-unmarshal-map` checker.

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
	for _, v := range resp.Violations {
		if v.Checker != "json-unmarshal-map" {
			t.Fatalf("expected checker 'json-unmarshal-map', got %q", v.Checker)
		}
	}
}
```
