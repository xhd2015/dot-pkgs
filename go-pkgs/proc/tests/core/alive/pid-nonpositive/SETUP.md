# Scenario

**Feature**: `Alive` is false for pid ≤ 0

```
AlivePID=0 (no inject) -> Alive(0, {}) -> false
# Assert also probes -3 via proc.Alive for pid<0
```

## Steps

1. Set `req.AlivePID=0` with inject disabled.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.AlivePID = 0
	req.AliveUseInject = false
	return nil
}
```
