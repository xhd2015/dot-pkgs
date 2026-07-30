# Scenario

**Feature**: `Options.Alive` inject is respected for positive pids

```
# first Run path: inject true for pid 50
AliveUseInject + AliveInject=true, AlivePID=50 -> Alive -> true
# Assert re-probes inject false for pid 51 via proc.Alive directly
```

## Steps

1. Set `req.AlivePID=50`, `AliveUseInject=true`, `AliveInject=true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.AlivePID = 50
	req.AliveUseInject = true
	req.AliveInject = true
	return nil
}
```
