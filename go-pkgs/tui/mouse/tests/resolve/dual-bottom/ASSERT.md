## Expected

- `err` is nil.
- `resp.Resolve.OK` true.
- `resp.Resolve.Hit.ID` is `"gen-commit-msg"`.
- `resp.Resolve.Kind` is `"bottom"`.
- `resp.Resolve.LocalY` is 4.

## Errors

- Missing bottom candidate or resolving to tag-next for bottom-relative gen click.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	r := resp.Resolve
	if !r.OK || r.Hit.ID != "gen-commit-msg" {
		t.Fatalf("dual-bottom: want gen-commit-msg, got OK=%v ID=%q Kind=%q",
			r.OK, r.Hit.ID, r.Kind)
	}
	if r.Kind != "bottom" || r.LocalY != 4 {
		t.Fatalf("dual-bottom: Kind=%q LocalY=%d, want bottom LocalY=4", r.Kind, r.LocalY)
	}
}
```
