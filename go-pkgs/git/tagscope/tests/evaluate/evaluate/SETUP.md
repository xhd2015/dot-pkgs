# Scenario

**Feature**: Evaluate applies per-scope gate rules and plans NextTag

```
CollectedTags + OwnedTreePair + HeadCommit -> Evaluate -> ChangePlan
```

## Steps

1. Set `req.Op` to `"evaluate"`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "evaluate"
	return nil
}
```