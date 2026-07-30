# Scenario

**Feature**: `Alive(pid, opts)` — non-positive pids false; inject for positive pids

```
pid<=0 -> Alive -> false
pid>0 + Options.Alive inject -> Alive -> inject result
```

## Preconditions

- Leaves set `req.AlivePID` and optionally `AliveUseInject` / `AliveInject`.
- No requirement to spawn live processes for P1.

## Steps

1. Grouping documents alive ops; leaves set `req.Op="alive"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "alive"
	return nil
}
```
