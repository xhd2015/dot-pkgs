# Scenario

**Feature**: missing root PID yields empty descendants (no error)

```
table without pid 999, root=999 -> Descendants -> [] (empty, no error)
```

## Steps

1. Set `req.Op` to `"descendants"`.
2. Fixture is a normal chain; `RootPID=999` absent from table.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "descendants"
	req.RootPID = 999
	req.MaxDepth = 16
	req.FixtureProcs = []FixtureProc{
		{PID: 1, PPID: 0, Cmd: "root"},
		{PID: 2, PPID: 1, Cmd: "child"},
	}
	return nil
}
```
