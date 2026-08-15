# Scenario

**Feature**: top anchor click on btn-a still resolves despite origin near 0

```
# --anchor=top pad=0; absY = ORIGIN + 3 (~3) -> HIT btn-a
fixture top + ORIGIN≈0 -> click -> HIT id=btn-a localY=3
```

## Steps

1. Set Action click, LocalY=3, WantHitID=btn-a (Anchor set by parent).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Action = "click"
	req.LocalY = 3
	req.ClickCol = 5
	req.WantHitID = "btn-a"
	return nil
}
```
