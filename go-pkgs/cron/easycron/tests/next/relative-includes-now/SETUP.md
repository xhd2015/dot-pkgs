# Scenario

**Feature**: every-1h fires at anchor when from=anchor

```
Next(every-1h, anchor=10:00, from=10:00) -> 10:00
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
	req.Expr = "every-1h"
	req.Anchor = "2026-08-24T10:00:00Z"
	req.From = "2026-08-24T10:00:00Z"
	return nil
}
```
