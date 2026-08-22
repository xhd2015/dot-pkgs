# Scenario

**Feature**: empty process list yields unknown idle and a soft warning

```
ListProcs empty for tty -> Capture -> Idle=nil, Summary Unknown=1, warning
```

## Steps

1. One session with a tty present in hierarchy.
2. Set `ListProcsMode=empty` so enrich sees no processes (not idle/busy fixture).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	sess := baseSession(1, "44444444-4444-4444-4444-444444444444", "unknown-sess", "/dev/ttys004", "Default")
	req.Windows = []snapshot.SnapshotWindow{
		oneSessionWindow(1, "U", 1, 1, "T", sess),
	}
	// No IdleTTYs/BusyTTYs; override ListProcs to empty after fixture apply.
	req.ListProcsMode = "empty"
	return nil
}
```
