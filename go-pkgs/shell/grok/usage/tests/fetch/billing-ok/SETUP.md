# Scenario

**Feature**: Billing endpoint succeeds with monthly limit

```
GetJSON(billing)=200 used=73 limit=100 -> Source=billing UsedPercent=73
```

## Steps

1. Set `FetchMode=billing-ok`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "billing-ok"
	return nil
}
```
