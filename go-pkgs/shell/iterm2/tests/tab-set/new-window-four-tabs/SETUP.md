# Scenario

**Feature**: four-tab set creates one window, three extra tabs, four ordered commands

```
TabSetSpec{Name=bots, Tabs×4}
  -> BuildTabSetNewWindowScript
  -> 1× create window, 3× create tab, write text cmd-a..cmd-d in order
```

## Steps

1. Set set name to `bots`.
2. Four tabs with stable IDs `a`..`d` and commands `cmd-a`..`cmd-d`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.TabSetName = "bots"
	req.Tabs = []TabSpecInput{
		{ID: "a", Name: "Alpha", Command: "cmd-a"},
		{ID: "b", Name: "Beta", Command: "cmd-b"},
		{ID: "c", Name: "Gamma", Command: "cmd-c"},
		{ID: "d", Name: "Delta", Command: "cmd-d"},
	}
	return nil
}
```
