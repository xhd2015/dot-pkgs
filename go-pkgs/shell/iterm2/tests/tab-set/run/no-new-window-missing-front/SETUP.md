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
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

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
