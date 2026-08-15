## Expected

- Two decisions in scope order: `sub/` then `sub/nested/`.
- `sub/` `NextTag` is `sub/v0.2.4`.
- `sub/nested/` `NextTag` is `sub/nested/v0.1.2`.
- Both `SkipReason` values are empty.

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Plan.Decisions) != 2 {
		t.Fatalf("Decisions len = %d, want 2", len(resp.Plan.Decisions))
	}
	if resp.Plan.Decisions[0].Scope.PathPrefix != "sub/" {
		t.Fatalf("Decisions[0] scope = %q, want sub/", resp.Plan.Decisions[0].Scope.PathPrefix)
	}
	if resp.Plan.Decisions[1].Scope.PathPrefix != "sub/nested/" {
		t.Fatalf("Decisions[1] scope = %q, want sub/nested/", resp.Plan.Decisions[1].Scope.PathPrefix)
	}
	sub := decisionFor(t, resp.Plan, "sub/")
	nested := decisionFor(t, resp.Plan, "sub/nested/")
	if sub.NextTag != "sub/v0.2.4" {
		t.Fatalf("sub NextTag = %q, want sub/v0.2.4", sub.NextTag)
	}
	if nested.NextTag != "sub/nested/v0.1.2" {
		t.Fatalf("nested NextTag = %q, want sub/nested/v0.1.2", nested.NextTag)
	}
}
```