# Scenario

**Feature**: smart run with empty find creates a new window

```
Find([]) + Mode Smart -> RunTabSet
  -> CreatedWindow=true; Exec receives create-window script with tab commands
```

## Steps

1. Two config tabs (`t1`/`cmd-a`, `t2`/`cmd-b`).
2. Find returns empty.
3. Mode smart (default).

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
		{ID: "t1", Name: "A", Command: "cmd-a"},
		{ID: "t2", Name: "B", Command: "cmd-b"},
	}
	req.FindSessions = nil
	return nil
}
```
