# Scenario

**Feature**: `ChildrenIndex` maps each PPID to sorted child PIDs

```
forking fixture table -> ChildrenIndex -> map[ppid][]pid ascending
```

## Steps

1. Set `req.Op` to `"children-index"`.
2. Load a small forest: parent 10 with children 30 and 20; parent 20 with 40.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "children-index"
	// Intentionally unsorted input order; children of 10 must sort as [20,30].
	req.FixtureProcs = []FixtureProc{
		{PID: 30, PPID: 10, Cmd: "b"},
		{PID: 10, PPID: 1, Cmd: "parent"},
		{PID: 40, PPID: 20, Cmd: "c"},
		{PID: 20, PPID: 10, Cmd: "a"},
	}
	return nil
}
```
