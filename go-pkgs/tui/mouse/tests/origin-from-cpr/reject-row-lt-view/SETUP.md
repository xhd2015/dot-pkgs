# Scenario

**Feature**: row1 < viewLines is rejected (live rule; not clamped to 0)

```
OriginFromCPR(9, 20) -> (0, false)
# must not treat stale probe as top-anchored origin 0
```

## Steps

1. Set Row1=9, ViewLines=20 (row smaller than painted frame height).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.Row1 = 9
	req.ViewLines = 20
	return nil
}
```
