# Scenario

**Feature**: a failed CLI resolution stops before the runner

```
Resolve -> error -> OpenConfig error
Run not called
```

## Steps

1. Set `Kind=dir` and `ResolveErr`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "dir"
	req.ResolveErr = "lookpath: code: not found"
	return nil
}
```
