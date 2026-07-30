# Scenario

**Feature**: depth-limited BFS descendants on linear chain 1→2→3

```
chain 1→2→3, root=1, maxDepth=1 -> Descendants -> [1,2] (not 3)
```

## Steps

1. Set `req.Op` to `"descendants"`.
2. Use three-node chain; `RootPID=1`, `MaxDepth=1`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "descendants"
	req.RootPID = 1
	req.MaxDepth = 1
	req.FixtureProcs = []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "root"},
		{PID: 2, PPID: 1, Cmd: "child"},
		{PID: 3, PPID: 2, Cmd: "grand"},
	}
	return nil
}
```
