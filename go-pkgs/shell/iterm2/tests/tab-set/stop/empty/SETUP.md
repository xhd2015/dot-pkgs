# Scenario

**Feature**: stop with no sessions returns not-running warning and zero closes

```
Find([]) -> StopTabSet -> err=nil, ClosedWindows=0, ClosedTabs=0, Warning non-empty
```

## Steps

1. Empty FindSessions.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.TabSetName = "bots"
	req.FindSessions = nil
	return nil
}
```
