# Scenario

**Feature**: mid-pane CPR after FrameSuffix becomes PhaseKnown with origin 6

```
FrameSuffix(26,20)=CSI6n -> Pending
OnCPR(26,1) -> Known, OriginY=6
```

## Steps

1. Probe with height 26, viewLines 20.
2. Apply CPR row1=26 col1=1 (cursor on last painted line).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.TrackerSteps = []TrackerStep{
		{Kind: "frame-suffix", Height: 26, ViewLines: 20},
		{Kind: "on-cpr", Row1: 26, Col1: 1},
	}
	return nil
}
```
