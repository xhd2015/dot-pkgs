# Scenario

**Feature**: OriginFromCPR derives live origin from CPR row and painted viewLines

```
# cursor on last line of frame: originY0 = row1 - viewLines
# live rule: reject when row1 < viewLines (stale must not look top-anchored)
row1, viewLines -> OriginFromCPR -> (originY0, ok)
```

## Preconditions

- `req.Op = "origin-from-cpr"`.
- Leaves set Row1 and ViewLines only (no clamped variant here).

## Steps

1. Set Op.
2. Leaf chooses valid mid-pane CPR or row1 < viewLines reject path.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "origin-from-cpr"
	return nil
}
```
