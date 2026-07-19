# Scenario

**Feature**: stop finds marked sessions and Exec's close scripts

```
Find(two sessions) -> StopTabSet
  -> Exec scripts mention close; ClosedWindows or ClosedTabs > 0
```

## Steps

1. Find two sessions in same window.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.TabSetName = "bots"
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-1", SessionID: "s1", TTY: "/dev/ttys050"},
		{TabID: "t2", WindowID: "win-1", SessionID: "s2", TTY: "/dev/ttys051"},
	}
	return nil
}
```
