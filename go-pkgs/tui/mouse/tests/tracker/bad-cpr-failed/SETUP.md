# Scenario

**Feature**: CPR with row1 < pending viewLines transitions to PhaseFailed

```
FrameSuffix(26,20) -> Pending
OnCPR(9,1) -> Failed, OriginY nil, OnCPR returns false
```

## Steps

1. Probe height 26, viewLines 20.
2. Apply bad CPR row1=9 (less than viewLines).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.TrackerSteps = []TrackerStep{
		{Kind: "frame-suffix", Height: 26, ViewLines: 20},
		{Kind: "on-cpr", Row1: 9, Col1: 1},
	}
	return nil
}
```
