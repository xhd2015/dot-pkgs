# Scenario

**Feature**: one window / one tab / one idle shell-only session

```
1 window + IdleTTY ttys001 -> Capture -> Summary Idle=1, session Idle=true
```

## Steps

1. Fixture: window 1 "Main", tab 1, session on `/dev/ttys001`.
2. Mark `ttys001` idle (shell only via phased fixture).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	sess := baseSession(1, "11111111-1111-1111-1111-111111111111", "idle-shell", "/dev/ttys001", "Default")
	req.Windows = []snapshot.SnapshotWindow{
		oneSessionWindow(1, "Main", 42, 1, "Tab1", sess),
	}
	req.IdleTTYs = []string{"ttys001"}
	return nil
}
```
