# Scenario

**Feature**: absY=9 with originY=6 hits add-changes (localY=3)

```
Resolve known: absY=9 -> LocalY=3 -> Hit.ID add-changes, Kind known
```

## Steps

1. Set AbsY to 9 (local row of add-changes chip).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AbsY = 9
	return nil
}
```
