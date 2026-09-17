# Scenario

**Feature**: reuse is the default window mode

```
Args("code", dir, false) -> ["code", "-r", dir]
```

## Steps

1. Set `Kind=dir`, `NewWindow=false`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "dir"
	req.NewWindow = false
	return nil
}
```
