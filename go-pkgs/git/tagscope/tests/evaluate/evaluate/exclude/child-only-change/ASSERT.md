## Expected

- Two decisions: `sub/` then `sub/nested/`.
- `sub/` `SkipReason` is `no-changes`, `NextTag` empty.
- `sub/nested/` `NextTag` is `sub/nested/v0.1.2`, `SkipReason` empty.

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
	sub := decisionFor(t, resp.Plan, "sub/")
	nested := decisionFor(t, resp.Plan, "sub/nested/")
	if sub.SkipReason != "no-changes" {
		t.Fatalf("sub SkipReason = %q, want no-changes", sub.SkipReason)
	}
	if sub.NextTag != "" {
		t.Fatalf("sub NextTag = %q, want empty", sub.NextTag)
	}
	if nested.NextTag != "sub/nested/v0.1.2" {
		t.Fatalf("nested NextTag = %q, want sub/nested/v0.1.2", nested.NextTag)
	}
	if nested.SkipReason != "" {
		t.Fatalf("nested SkipReason = %q, want empty", nested.SkipReason)
	}
}
```