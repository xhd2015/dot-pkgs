# Scenario

**Feature**: expressions without not-within are always active

```
Active(any) == true for every-1h
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
	req.At = "2026-08-24T19:01:00Z"
	return nil
}
```
