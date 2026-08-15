## Expected

- `err` is nil.
- `resp.Resolve.OK` true.
- `resp.Resolve.Hit.ID` is `"add-changes"`.
- `resp.Resolve.Kind` is `"known"`.
- `resp.Resolve.LocalY` is 3.

## Errors

- Wrong chip (gen/tag) or dual Kind when origin is known.

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
	if !r.OK || r.Hit.ID != "add-changes" || r.Kind != "known" || r.LocalY != 3 {
		t.Fatalf("Resolve absY=9: got OK=%v ID=%q Kind=%q LocalY=%d, want add-changes known LocalY=3",
			r.OK, r.Hit.ID, r.Kind, r.LocalY)
	}
}
```
