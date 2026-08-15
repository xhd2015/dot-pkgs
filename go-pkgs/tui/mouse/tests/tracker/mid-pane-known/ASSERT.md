## Expected

- `err` is nil.
- `resp.LastSuffix` is CSI6n (`\x1b[6n`).
- `resp.LastOnCPR` is true.
- `resp.Phase` is `PhaseKnown`.
- `resp.TrackerOriginY` non-nil and equal to 6.

## Errors

- Staying Pending/Failed or wrong origin after a valid mid-pane CPR.

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
	if resp.LastSuffix != mouse.CSI6n {
		t.Fatalf("FrameSuffix=%q, want CSI6n", resp.LastSuffix)
	}
	if !resp.LastOnCPR {
		t.Fatal("OnCPR returned false, want true for mid-pane CPR")
	}
	if resp.Phase != mouse.PhaseKnown {
		t.Fatalf("Phase=%v, want PhaseKnown", resp.Phase)
	}
	if resp.TrackerOriginY == nil || *resp.TrackerOriginY != 6 {
		t.Fatalf("OriginY=%v, want 6", resp.TrackerOriginY)
	}
}
```
