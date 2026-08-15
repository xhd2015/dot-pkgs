# Scenario

**Feature**: after mid paint, origin is locked and/or host cursor is on last UI line

```
# paint mid UI + CSI 6n; CPR -> ORIGIN; query-cursor reads host VT cursor
fixture --anchor=mid paint -> ORIGIN=n VIEW=5
exec tty-watch send --query-cursor --json -> row/col near last UI line
```

## Steps

1. Set Action to query-cursor (no click).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Action = "query-cursor"
	return nil
}
```
