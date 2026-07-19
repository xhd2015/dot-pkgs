# Scenario

**Feature**: multi-window find sets Warning and syncs only the first (most recent) window

```
Find(win-A then win-B) -> RunTabSet
  -> Warning mentions 2 windows; FocusedWindow is first window win-A
```

## Steps

1. Two sessions, different WindowIDs, same set (order: win-A first = most recent).
2. Both idle; config has both tabs.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunMode = "smart"
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "A", Command: "cmd-a"},
		{ID: "t2", Name: "B", Command: "cmd-b"},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-A", SessionID: "s1", TTY: "/dev/ttys020"},
		{TabID: "t2", WindowID: "win-B", SessionID: "s2", TTY: "/dev/ttys021"},
	}
	req.BusyByTab = map[string]string{"t1": "idle", "t2": "idle"}
	return nil
}
```
