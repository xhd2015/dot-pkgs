# Scenario

**Feature**: session Cwd is attached from ListCwds for the session tty

```
IdleTTY + CwdByTTY[ttys003]=/Users/me/proj -> Capture -> session.Cwd that path
```

## Steps

1. One idle session on `/dev/ttys003`.
2. Set `CwdByTTY` so phased ListCwds returns `/Users/me/proj` for that tty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	sess := baseSession(1, "33333333-3333-3333-3333-333333333333", "cwd-sess", "/dev/ttys003", "Default")
	req.Windows = []snapshot.SnapshotWindow{
		oneSessionWindow(1, "CwdWin", 7, 1, "CwdTab", sess),
	}
	req.IdleTTYs = []string{"ttys003"}
	req.CwdByTTY = map[string]string{"ttys003": "/Users/me/proj"}
	return nil
}
```
