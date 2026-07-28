# Scenario

**Feature**: create missing tab in window stages command when NoSubmit

```
Find(only t1) + config {t1, t2 NoSubmit} -> RunTabSet
  -> t2 Action created; create-tab script write text "missing-staged" without newline
```

## Steps

1. Config tabs `t1` (present, idle) and `t2` (missing, command `missing-staged`, NoSubmit=true).
2. Find only returns `t1`.

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
		{ID: "t1", Name: "One", Command: "cmd-one"},
		{ID: "t2", Name: "Two", Command: "missing-staged", NoSubmit: true},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-1", SessionID: "s1", TTY: "/dev/ttys022"},
	}
	req.BusyByTab = map[string]string{"t1": "idle"}
	return nil
}
```
