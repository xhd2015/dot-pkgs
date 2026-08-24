# Scenario

**Feature**: every-1h-at-4m from 10:07 snaps to 11:04

```
Next(every-1h-at-4m, from=10:07) -> 11:04
```

## Steps

1. Configure the request fields below.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Expr = "every-1h-at-4m"
	req.Anchor = "2026-08-24T10:07:00Z"
	req.From = "2026-08-24T10:07:00Z"
	return nil
}
```
