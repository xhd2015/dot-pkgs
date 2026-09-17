# Scenario

**Feature**: `NewWindow` selects `-n`

```
Args("code", dir, true) -> ["code", "-n", dir]
```

## Steps

1. Set `Kind=dir`, `NewWindow=true`.

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
