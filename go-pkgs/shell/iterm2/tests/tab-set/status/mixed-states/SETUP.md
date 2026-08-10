# Scenario

**Feature**: status maps busy→running, idle→idle, unknown→unknown, absent→missing

```
config tabs t-run,t-idle,t-unk,t-miss + Find three of them
  -> StatusTabSet States running/idle/unknown/missing; Instances >= 1
```

## Steps

1. Four config tabs; find three with different busy states; one missing.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t-run", Name: "Run", Command: "c1"},
		{ID: "t-idle", Name: "Idle", Command: "c2"},
		{ID: "t-unk", Name: "Unk", Command: "c3"},
		{ID: "t-miss", Name: "Miss", Command: "c4"},
	}
	req.FindSessions = []SessionRefInput{
		{TabID: "t-run", WindowID: "win-1", TTY: "/dev/ttys040"},
		{TabID: "t-idle", WindowID: "win-1", TTY: "/dev/ttys041"},
		{TabID: "t-unk", WindowID: "win-1", TTY: "/dev/ttys042"},
	}
	req.BusyByTab = map[string]string{
		"t-run":  "busy",
		"t-idle": "idle",
		"t-unk":  "unknown",
	}
	return nil
}
```
