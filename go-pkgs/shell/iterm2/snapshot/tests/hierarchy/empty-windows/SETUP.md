# Scenario

**Feature**: empty window list yields empty snapshot with zero summary counts

```
Windows=[] + ITermRunning=true -> Capture -> Windows empty, Summary all zeros
```

## Steps

1. Keep iTerm running from parent.
2. Leave `Windows` empty (no tabs/sessions).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Windows = nil
	req.IdleTTYs = nil
	req.BusyTTYs = nil
	return nil
}
```
