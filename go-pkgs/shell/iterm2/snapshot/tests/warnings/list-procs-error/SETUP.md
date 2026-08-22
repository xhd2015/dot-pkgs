# Scenario

**Feature**: ListProcs error becomes a soft warning; Capture still succeeds

```
ListProcs returns error -> Capture -> err=nil, warnings mention ps failure
```

## Steps

1. One session on `/dev/ttys005`.
2. `ListProcsMode=error` with message `ps failed for fixture`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	sess := baseSession(1, "55555555-5555-5555-5555-555555555555", "warn-sess", "/dev/ttys005", "Default")
	req.Windows = []snapshot.SnapshotWindow{
		oneSessionWindow(1, "W", 1, 1, "T", sess),
	}
	req.ListProcsMode = "error"
	req.ListProcsErr = "ps failed for fixture"
	return nil
}
```
