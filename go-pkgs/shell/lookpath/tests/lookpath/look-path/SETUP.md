# Scenario

**Feature**: `LookPath(name, opts)` convenience returns path string only

```
LookPath == Look(...).Path (same pipeline injectables)
```

## Steps

1. Set `Operation=look-path`.
2. Leaves configure hit vs miss fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "look-path"
	req.Name = "mytool"
	return nil
}
```
