# Scenario

**Feature**: orchestration never auto-sends Ctrl+C when resending or creating

```
idle resend path -> Exec scripts must not contain control-c / ctrl keystroke c
```

## Steps

1. Smart run with idle tab (resend path) to force mutating scripts.
2. Assert scans all captured Exec scripts.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunMode = "smart"
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "A", Command: "safe-resend"},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t1", WindowID: "win-1", SessionID: "s1", TTY: "/dev/ttys030"},
	}
	req.BusyByTab = map[string]string{"t1": "idle"}
	return nil
}
```
