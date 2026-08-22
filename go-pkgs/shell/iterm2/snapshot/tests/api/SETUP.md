# Scenario

**Feature**: public Capture API surface on injectable Collector

```
NewCollector + injects -> Capture | CaptureWith(zero) -> equivalent snapshot
```

## Steps

1. iTerm running for API leaves.
2. Leaves select Capture vs CaptureWith.

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
