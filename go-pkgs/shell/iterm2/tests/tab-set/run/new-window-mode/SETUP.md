# Scenario

**Feature**: NewWindow mode always creates even when find has sessions

```
Find(non-empty) + Mode NewWindow -> RunTabSet
  -> always create window script; does not only focus existing
```

## Steps

1. Find returns an existing session for `t1`.
2. Mode `new-window`.
3. Config still has `t1`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunMode = "new-window"
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "A", Command: "force-new-cmd"},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-old", SessionID: "s1", TTY: "/dev/ttys001"},
	}
	req.BusyByTab = map[string]string{"t1": "idle"}
	return nil
}
```
