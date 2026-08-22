# Scenario

**Feature**: process enrichment classifies idle/busy/unknown and attaches cwd

```
ListProcs/ListCwds inject -> Capture enrich phase -> Idle/Busy/Unknown + Cwd
```

## Steps

1. Grouping requires iTerm running.
2. Leaves set tty enrich modes and cwd maps.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermRunning = true
	return nil
}
```
