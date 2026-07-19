# Scenario

**Feature**: click inside the run chip rectangle returns ID "run"

```
Hits left|run; (x=65, localY=3) -> HitTest -> Hit.ID == "run", ok
```

## Steps

1. Point at x=65, localY=3 (inside run: X0=61,X1=71, Y0=3,Y1=4).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.X = 65
	req.LocalY = 3
	return nil
}
```
