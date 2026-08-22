## Expected

- CaptureWith with zero opts succeeds like Capture.
- Summary Idle=1 for the idle fixture; Host/Source set from injects.
- Demonstrates parallel-safe instance inject (no process-global collector).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if !req.UseCaptureWith {
		t.Fatal("this leaf must exercise CaptureWith")
	}
	snap := mustSnap(t, resp, err)
	assertSummary(t, snap.Summary, snapshot.SnapshotSummary{
		Windows: 1, Tabs: 1, Sessions: 1, Idle: 1, Busy: 0, Unknown: 0,
	})
	if snap.Source != "iterm2" {
		t.Fatalf("Source=%q", snap.Source)
	}
	if snap.Host != "testhost" {
		t.Fatalf("Host=%q", snap.Host)
	}
	idle, ok := boolVal(snap.Windows[0].Tabs[0].Sessions[0].Idle)
	if !ok || !idle {
		t.Fatalf("want idle via CaptureWith, got %#v", snap.Windows[0].Tabs[0].Sessions[0].Idle)
	}
}
```
