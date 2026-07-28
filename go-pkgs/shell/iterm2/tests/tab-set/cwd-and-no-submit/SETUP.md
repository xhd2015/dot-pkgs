# Scenario

**Feature**: cwd cd always executes (newline); command honors NoSubmit

```
TabSpec{Cwd: /tmp/stage-cwd, Command: staged-cwd-cmd, NoSubmit: true}
  -> cd write still with newline / expression form (executes)
  -> command write text "…" without newline
```

## Steps

1. One tab with non-empty Cwd, command `staged-cwd-cmd`, NoSubmit=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.TabSetName = "cwd-stage"
	req.Tabs = []TabSpecInput{
		{
			ID:       "c1",
			Name:     "CwdStage",
			Command:  "staged-cwd-cmd",
			Cwd:      "/tmp/stage-cwd",
			NoSubmit: true,
		},
	}
	return nil
}
```
