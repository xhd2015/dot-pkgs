# Scenario

**Feature**: Evaluate applies per-scope gate rules and plans NextTag

```
CollectedTags + OwnedTreePair + HeadCommit -> Evaluate -> ChangePlan
```

## Steps

1. Set `req.Op` to `"evaluate"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "evaluate"
	return nil
}
```