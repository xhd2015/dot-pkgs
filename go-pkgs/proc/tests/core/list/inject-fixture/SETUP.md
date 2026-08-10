# Scenario

**Feature**: custom `Options.List` return value is used as-is by `List`

```
ListInject two rows -> List(Options{List:…}) -> same two rows
```

## Steps

1. Set `req.ListInject` to a two-process fixture.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ListInject = []FixtureProc{
		{PID: 11, PPID: 1, Cmd: "fixture-a"},
		{PID: 22, PPID: 11, Cmd: "fixture-b --flag"},
	}
	return nil
}
```
