# Scenario

**Feature**: stop with no sessions returns not-running warning and zero closes

```
Find([]) -> StopTabSet -> err=nil, ClosedWindows=0, ClosedTabs=0, Warning non-empty
```

## Steps

1. Empty FindSessions.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d

	req.TabSetName = "bots"
	req.FindSessions = nil
	return nil
}
```
