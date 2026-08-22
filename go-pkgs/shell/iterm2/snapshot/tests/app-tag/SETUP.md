# Scenario

**Feature**: AppTag stamps empty Window.App on capture

```
fixture window App="" + Collector.AppTag -> output Window.App == AppTag
```

## Steps

1. Leaves set AppTag and a simple idle window with empty App.

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
