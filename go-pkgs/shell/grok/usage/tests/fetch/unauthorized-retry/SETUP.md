# Scenario

**Feature**: Billing 401 then force-refresh retry succeeds

```
GetJSON=401 -> Ensure(ForceRefresh) -> GetJSON=200
```

## Steps

1. Set `FetchMode=unauthorized-retry`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "unauthorized-retry"
	return nil
}
```
