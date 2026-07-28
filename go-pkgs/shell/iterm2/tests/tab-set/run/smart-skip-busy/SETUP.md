# Scenario

**Feature**: smart run skips busy tabs without resending command

```
Find(t1 busy) + Mode Smart -> RunTabSet
  -> Action skipped-busy; Exec scripts must not write text busy-cmd
```

## Steps

1. Config tab `t1` command `busy-cmd`.
2. Find has `t1` in `win-1`.
3. Busy map: `t1` → busy.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.RunMode = "smart"
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "Busy", Command: "busy-cmd"},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-1", SessionID: "s1", TTY: "/dev/ttys010"},
	}
	req.BusyByTab = map[string]string{"t1": "busy"}
	return nil
}
```
