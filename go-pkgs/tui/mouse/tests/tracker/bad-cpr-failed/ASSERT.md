## Expected

- `err` is nil.
- `resp.LastOnCPR` is false.
- `resp.Phase` is `PhaseFailed`.
- `resp.TrackerOriginY` is nil.

## Errors

- Accepting bad CPR as Known or leaving Pending without Failed.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("tracker: %v", err)
	}
	if resp.LastOnCPR {
		t.Fatal("OnCPR returned true for row1 < viewLines")
	}
	if resp.Phase != mouse.PhaseFailed {
		t.Fatalf("Phase=%v, want PhaseFailed", resp.Phase)
	}
	if resp.TrackerOriginY != nil {
		t.Fatalf("OriginY=%v, want nil after Failed", *resp.TrackerOriginY)
	}
}
```
