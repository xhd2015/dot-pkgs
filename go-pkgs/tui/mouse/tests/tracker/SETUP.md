# Scenario

**Feature**: Tracker owns CSI 6n origin measurement phase for inline TUI

```
# Unknown --FrameSuffix--> Pending --OnCPR good--> Known
#                          Pending --OnCPR bad---> Failed
# Known --OnResize--> Unknown (re-probe)
Tracker steps -> Phase + OriginY
```

## Preconditions

- `req.Op = "tracker"`.
- Leaves supply ordered `TrackerSteps` (frame-suffix / on-cpr / on-resize).

## Steps

1. Set Op to tracker.
2. Leaf defines the step sequence and asserts final Phase / OriginY.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Op = "tracker"
	return nil
}
```
