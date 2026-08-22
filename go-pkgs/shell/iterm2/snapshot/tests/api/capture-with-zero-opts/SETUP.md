# Scenario

**Feature**: CaptureWith(zero CaptureOpts) matches Capture behavior on injects

```
UseCaptureWith=true + idle fixture -> CaptureWith({}) -> same summary as Capture
```

## Steps

1. Same single-idle fixture as hierarchy/single-idle.
2. Set `UseCaptureWith=true` so Run calls `CaptureWith(CaptureOpts{})`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	sess := baseSession(1, "66666666-6666-6666-6666-666666666666", "api-idle", "/dev/ttys006", "Default")
	req.Windows = []snapshot.SnapshotWindow{
		oneSessionWindow(1, "API", 3, 1, "T", sess),
	}
	req.IdleTTYs = []string{"ttys006"}
	req.UseCaptureWith = true
	return nil
}
```
