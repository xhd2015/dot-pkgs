## Expected

- `err` is nil.
- `resp.Resolve.OK` true.
- `resp.Resolve.Hit.ID` is `"gen-commit-msg"`.
- `resp.Resolve.Kind` is `"known"`.
- `resp.Resolve.LocalY` is 4.

## Errors

- Resolving to add-changes or tag-next for this absY.

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
	if !r.OK || r.Hit.ID != "gen-commit-msg" || r.Kind != "known" || r.LocalY != 4 {
		t.Fatalf("Resolve absY=10: got OK=%v ID=%q Kind=%q LocalY=%d, want gen-commit-msg known LocalY=4",
			r.OK, r.Hit.ID, r.Kind, r.LocalY)
	}
}
```
