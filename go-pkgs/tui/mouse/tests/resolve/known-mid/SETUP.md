# Scenario

**Feature**: known mid-pane origin (originY=6) resolves Run chips by absY

```
OriginY=6, Height=26, ViewLines=20, midPaneHits
localY = absY - 6  # absY 9 → 3 (add), absY 10 → 4 (gen)
```

## Preconditions

- Origin is Known mid-pane: pointer to 6.
- Hits are add-changes / gen-commit-msg / tag-next stack.
- Height 26, ViewLines 20 (same paint geometry as unit fixture).

## Steps

1. Install mid-pane hits and known origin 6.
2. Leaf chooses AbsY 9 or 10 on the run X column.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	oy := 6
	req.OriginY = &oy
	req.Hits = midPaneHits()
	req.Height = 26
	req.ViewLines = 20
	req.AbsX = 67
	return nil
}
```
