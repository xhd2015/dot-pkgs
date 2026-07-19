# Scenario

**Feature**: localY equal to y1 is outside the half-open interval

```
Hits left|run (Y0=3,Y1=4); localY=4 -> HitTest -> ok=false
```

## Steps

1. Keep run-row hits; set localY to the exclusive upper bound (4) and x still on run.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.X = 65
	req.LocalY = 4
	return nil
}
```
