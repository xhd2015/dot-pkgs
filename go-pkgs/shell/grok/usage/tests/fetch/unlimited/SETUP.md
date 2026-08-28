# Scenario

**Feature**: Billing succeeds with monthlyLimit 0 (unknown percent)

```
GetJSON(billing)=200 used=73 limit=0 -> percents=-1
```

## Steps

1. Set `FetchMode=unlimited`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "unlimited"
	return nil
}
```
