# Scenario

**Feature**: smart run recreates config tabs missing from find

```
Find(only t1) + config {t1,t2} -> RunTabSet
  -> t2 Action created; Exec contains create tab and missing-cmd
```

## Steps

1. Config tabs `t1` (present, idle) and `t2` (missing, command `missing-cmd`).
2. Find only returns `t1`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.RunMode = "smart"
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "One", Command: "cmd-one"},
		{ID: "t2", Name: "Two", Command: "missing-cmd"},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-1", SessionID: "s1", TTY: "/dev/ttys012"},
	}
	req.BusyByTab = map[string]string{"t1": "idle"}
	return nil
}
```
