# Scenario

**Feature**: row1=26 with viewLines=20 yields originY0=6

```
OriginFromCPR(26, 20) -> (6, true)
```

## Steps

1. Set Row1=26, ViewLines=20 (mid-pane paint: cursor on last of 20 lines).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Row1 = 26
	req.ViewLines = 20
	return nil
}
```
