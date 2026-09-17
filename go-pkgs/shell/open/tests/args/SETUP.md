# Scenario

**Feature**: pure argv builders — no process is launched

```
Args / URLArgs / AppArgs -> argv or error
```

## Steps

1. Set `Operation=args`.
2. Leaves set `Kind` and the target fields for that argv shape.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "args"
	return nil
}
```
