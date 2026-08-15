## Expected

- `err` is nil.
- `resp.Resolve.OK` true.
- `resp.Resolve.Hit.ID` is `"gen-commit-msg"` (not `"tag-next"`).
- `resp.Resolve.Kind` is `"top"`.
- `resp.Resolve.LocalY` is 4.

## Errors

- Mis-resolving to tag-next when the top-anchored gen row was clicked.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	r := resp.Resolve
	if !r.OK || r.Hit.ID != "gen-commit-msg" {
		t.Fatalf("dual-top: want gen-commit-msg, got OK=%v ID=%q Kind=%q",
			r.OK, r.Hit.ID, r.Kind)
	}
	if r.Kind != "top" || r.LocalY != 4 {
		t.Fatalf("dual-top: Kind=%q LocalY=%d, want top LocalY=4", r.Kind, r.LocalY)
	}
}
```
