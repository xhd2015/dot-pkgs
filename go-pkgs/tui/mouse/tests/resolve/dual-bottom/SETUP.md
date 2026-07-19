# Scenario

**Feature**: dual-origin bottom-anchored relative gen click still resolves gen

```
# BottomOriginY(40,20)=20; gen localY=4 => absY=24
OriginY nil, AbsY=24 -> bottom candidate LocalY=4 -> gen-commit-msg Kind=bottom
# top candidate absY=24 misses (no hit at y=24)
```

## Preconditions

- `OriginY` is nil.
- Same dual gen/tag hits as dual-top.
- Height 40, ViewLines 20 → bottom origin 20.

## Steps

1. Install dual hits; set AbsY so bottom localY lands on gen row (4+20=24).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.OriginY = nil
	req.Hits = dualGenTagHits()
	req.Height = 40
	req.ViewLines = 20
	req.AbsX = 65
	// bottom localY = absY - (height-viewLines) = 24 - 20 = 4 → gen
	req.AbsY = 24
	return nil
}
```
