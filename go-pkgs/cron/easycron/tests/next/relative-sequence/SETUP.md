# Scenario

**Feature**: every-1h next after anchor is +1h

```
Next(every-1h, from=anchor+1ns) -> anchor+1h
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
	req.From = "2026-08-24T10:00:00.000000001Z"
	return nil
}
```
