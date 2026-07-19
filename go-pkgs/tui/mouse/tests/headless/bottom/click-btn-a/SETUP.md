# Scenario

**Feature**: bottom anchor click on btn-a still resolves with high origin

```
# --anchor=bottom large pad; absY = ORIGIN + 3 -> HIT btn-a
fixture bottom + ORIGIN high -> click -> HIT id=btn-a localY=3
```

## Steps

1. Set Action click, LocalY=3, WantHitID=btn-a (Anchor set by parent).

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
