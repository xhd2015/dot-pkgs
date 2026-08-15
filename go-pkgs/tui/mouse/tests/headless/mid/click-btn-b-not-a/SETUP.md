# Scenario

**Feature**: mid click on btn-b (localY=4) resolves btn-b, not btn-a

```
# absY = ORIGIN + 4 -> HIT btn-b; must not be btn-a for this click
fixture mid -> click localY=4 col=5 -> HIT id=btn-b localY=4
```

## Steps

1. Set Action click, LocalY=4, WantHitID=btn-b.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Action = "click"
	req.LocalY = 4
	req.ClickCol = 5
	req.WantHitID = "btn-b"
	return nil
}
```
