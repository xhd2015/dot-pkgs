# Scenario

**Feature**: `List(opts)` uses `Options.List` inject when set

```
Options.List = fixture hook -> List(opts) -> fixture []Proc (no live ps)
```

## Preconditions

- Leaves set `req.ListInject` and `req.Op="list"`.
- Production live `ps` path is not required for P1 exit.

## Steps

1. Set `req.Op` to `"list"`.
2. Leaf provides inject rows.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "list"
	return nil
}
```
