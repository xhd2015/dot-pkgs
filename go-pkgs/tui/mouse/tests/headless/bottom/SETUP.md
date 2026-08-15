# Scenario

**Feature**: bottom anchor (large pad) for headless mouse geometry

```
# bottom: pad ≈ height-VIEW; ORIGIN high; click still maps to btn-a
fixture --anchor=bottom -> ORIGIN near bottom -> click localY=3 -> HIT btn-a
```

## Steps

1. Set `req.Anchor = "bottom"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Anchor = "bottom"
	return nil
}
```
