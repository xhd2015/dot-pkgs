# Scenario

**Feature**: `NewWindow` reaches the runner

```
Resolve -> Run(["code", "-n", dir])
```

## Steps

1. Set `Kind=dir` and `NewWindow=true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "dir"
	req.NewWindow = true
	return nil
}
```
