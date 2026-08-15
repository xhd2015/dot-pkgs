# Scenario

**Feature**: dual-origin top-anchored click on gen row is gen-commit-msg, not tag-next

```
# OriginY nil: try top localY=absY first
AbsY=4 (gen local row), Height=40 ViewLines=20
-> Resolve Kind=top Hit gen-commit-msg (must not be tag-next)
```

## Preconditions

- `OriginY` is nil (dual mode).
- Hits are gen-commit-msg (Y=4) and tag-next (Y=11).
- Height 40, ViewLines 20 so bottom origin is 20 (unused for this click).

## Steps

1. Install dual gen/tag hits with nil origin.
2. Click abs (65, 4) on the gen chip under top anchoring.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.OriginY = nil
	req.Hits = dualGenTagHits()
	req.Height = 40
	req.ViewLines = 20
	req.AbsX = 65
	req.AbsY = 4
	return nil
}
```
