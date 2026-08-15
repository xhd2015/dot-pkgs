## Expected

- Two decisions in `Collected.Scopes` order: root then `sub/`.
- Root `NextTag` is `v0.0.3`; `sub/` `NextTag` is `sub/v0.2.4`.
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
	if resp.Plan.Decisions[0].Scope.PathPrefix != "" {
		t.Fatalf("Decisions[0] scope = %q, want root", resp.Plan.Decisions[0].Scope.PathPrefix)
	}
	if resp.Plan.Decisions[1].Scope.PathPrefix != "sub/" {
		t.Fatalf("Decisions[1] scope = %q, want sub/", resp.Plan.Decisions[1].Scope.PathPrefix)
	}
	root := decisionFor(t, resp.Plan, "")
	sub := decisionFor(t, resp.Plan, "sub/")
	if root.NextTag != "v0.0.3" {
		t.Fatalf("root NextTag = %q, want v0.0.3", root.NextTag)
	}
	if sub.NextTag != "sub/v0.2.4" {
		t.Fatalf("sub NextTag = %q, want sub/v0.2.4", sub.NextTag)
	}
	if root.SkipReason != "" || sub.SkipReason != "" {
		t.Fatalf("SkipReason = [%q, %q], want empty", root.SkipReason, sub.SkipReason)
	}
}
```