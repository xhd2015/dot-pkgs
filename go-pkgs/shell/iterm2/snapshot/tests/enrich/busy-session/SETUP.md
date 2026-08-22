# Scenario

**Feature**: non-shell child process classifies session as busy

```
BusyTTY ttys002 + leaf "python train.py" -> Capture -> Idle=false, Summary Busy=1
```

## Steps

1. One session on `/dev/ttys002` marked busy via phased fixture.
2. Optional leaf command override `python train.py`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	sess := baseSession(1, "22222222-2222-2222-2222-222222222222", "busy-work", "/dev/ttys002", "Default")
	req.Windows = []snapshot.SnapshotWindow{
		oneSessionWindow(1, "Work", 99, 1, "Run", sess),
	}
	req.BusyTTYs = []string{"ttys002"}
	req.BusyLeafByTTY = map[string]string{"ttys002": "python train.py"}
	return nil
}
```
