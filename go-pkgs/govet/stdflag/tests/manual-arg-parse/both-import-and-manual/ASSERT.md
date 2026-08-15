## Expected
- Two violations: one from `std-flag` (import detection) and one from `manual-flag-parse` (manual parsing).

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
	hasStdFlag := false
	hasManual := false
	for _, v := range resp.Violations {
		if v.Checker == "std-flag" {
			hasStdFlag = true
		}
		if v.Checker == "manual-flag-parse" {
			hasManual = true
		}
	}
	if !hasStdFlag {
		t.Fatalf("expected a 'std-flag' violation for import flag, but none found")
	}
	if !hasManual {
		t.Fatalf("expected a 'manual-flag-parse' violation, but none found")
	}
}
```
