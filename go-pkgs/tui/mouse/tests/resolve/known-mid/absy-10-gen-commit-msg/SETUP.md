# Scenario

**Feature**: absY=10 with originY=6 hits gen-commit-msg (localY=4)

```
Resolve known: absY=10 -> LocalY=4 -> Hit.ID gen-commit-msg, Kind known
```

## Steps

1. Set AbsY to 10 (local row of gen-commit-msg chip).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	t.Helper()
	req.AbsY = 10
	return nil
}
```
