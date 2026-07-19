# Scenario

**Feature**: NoNewWindow without a frontmost window returns ErrNoITermWindow

```
Mode NoNewWindow + FrontmostWindowID="" + Find empty
  -> RunTabSet error is ErrNoITermWindow
```

## Steps

1. Mode `no-new-window`.
2. Empty Find and empty FrontmostWindowID.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RunMode = "no-new-window"
	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "t1", Name: "A", Command: "cmd-a"},
	}
	req.FindSessions = nil
	req.FrontmostWindowID = ""
	return nil
}
```
