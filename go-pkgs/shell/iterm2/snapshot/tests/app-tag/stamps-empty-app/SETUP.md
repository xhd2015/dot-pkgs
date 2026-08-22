# Scenario

**Feature**: AppTag fills empty App on deep-captured window

```
AppTag=~/Applications/iTerm.app + idle fixture -> Window.App stamped
```

## Steps

1. One idle window, App empty.
2. req.AppTag set to home install path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	sess := baseSession(1, "cccccccc-cccc-cccc-cccc-cccccccccccc", "shell", "/dev/ttys003", "Default")
	req.Windows = []snapshot.SnapshotWindow{
		oneSessionWindow(1, "HomeApp", 99, 1, "T1", sess),
	}
	req.IdleTTYs = []string{"ttys003"}
	req.AppTag = "~/Applications/iTerm.app"
	return nil
}
```
