# Scenario

**Feature**: soft warning paths during process enrich without failing Capture

```
ListProcs error -> Capture continues -> warnings[] non-empty, snapshot returned
```

## Steps

1. iTerm running; hierarchy present.
2. Leaves force soft probe failures.

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
