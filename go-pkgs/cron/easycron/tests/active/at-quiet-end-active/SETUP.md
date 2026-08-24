# Scenario

**Feature**: 06:30 is active (quiet end inclusive resume)

```
Active(06:30) == true
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
	req.Expr = "every-5m-not-within-19h00m-to-06h30m"
	req.At = "2026-08-25T06:30:00Z"
	return nil
}
```
