# Scenario

**Feature**: mid anchor click on btn-a (localY=3) resolves HIT id=btn-a

```
# absY = ORIGIN + 3, absX = 5 -> Resolve known -> HIT btn-a
fixture mid + ORIGIN -> send click -> HIT id=btn-a localY=3 kind=known
```

## Steps

1. Set Action click, LocalY=3, WantHitID=btn-a.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Action = "click"
	req.LocalY = 3
	req.ClickCol = 5
	req.WantHitID = "btn-a"
	return nil
}
```
