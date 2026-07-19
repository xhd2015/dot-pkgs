# Scenario

**Feature**: smart run resends command to idle tabs via write text only

```
Find(t1 idle) -> RunTabSet -> Action resent; Exec write text "idle-cmd"; no Ctrl+C
```

## Steps

1. Config tab `t1` command `idle-cmd`.
2. Find has idle `t1`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunMode = "smart"
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "Idle", Command: "idle-cmd"},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-1", SessionID: "s1", TTY: "/dev/ttys011"},
	}
	req.BusyByTab = map[string]string{"t1": "idle"}
	return nil
}
```
