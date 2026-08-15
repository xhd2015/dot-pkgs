## Expected

- `err` is nil.
- After resize + re-probe: `resp.LastSuffix` is CSI6n (query re-emitted).
- Final `resp.Phase` is `PhasePending` (Unknown → FrameSuffix).
- `resp.TrackerOriginY` is nil (origin cleared by OnResize; not yet Known again).

## Errors

- Keeping Known origin across resize, or failing to re-emit CSI6n.

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
		t.Fatalf("after resize FrameSuffix=%q, want CSI6n re-emit", resp.LastSuffix)
	}
	if resp.Phase != mouse.PhasePending {
		t.Fatalf("Phase=%v after resize re-probe, want PhasePending", resp.Phase)
	}
	if resp.TrackerOriginY != nil {
		t.Fatalf("OriginY=%v after resize, want nil until new CPR", *resp.TrackerOriginY)
	}
}
```
