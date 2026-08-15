# Scenario

**Feature**: OnResize invalidates Known origin back to Unknown and re-probes

```
FrameSuffix + good OnCPR -> Known
OnResize -> Unknown, OriginY nil
FrameSuffix again -> CSI6n (re-emit)
```

## Steps

1. Reach Known with mid-pane CPR (height 40, viewLines 20, row1=40).
2. Call OnResize.
3. Call FrameSuffix again; assert re-probe string and Unknown phase after resize
   (final phase is Pending after second FrameSuffix — assert suffix + that
   OriginY stayed nil through resize by checking LastSuffix re-emit and that
   we did not keep Known without re-probe).

Note: after the second FrameSuffix the phase is Pending. Assert focuses on
re-emitting CSI6n and that the sequence after resize is not still Known with
old origin (Run records final Phase after all steps = Pending).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.TrackerSteps = []TrackerStep{
		{Kind: "frame-suffix", Height: 40, ViewLines: 20},
		{Kind: "on-cpr", Row1: 40, Col1: 1},
		{Kind: "on-resize"},
		{Kind: "frame-suffix", Height: 40, ViewLines: 20},
	}
	return nil
}
```
