# Scenario

**Feature**: Capture fails hard when iTerm2 is not running

```
ITermRunning=false -> Capture -> error (no snapshot)
```

## Steps

1. Set `ITermRunning=false` (root default) with empty windows.
2. Call Capture via Run.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermRunning = false
	req.Windows = nil
	return nil
}
```
