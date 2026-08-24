# Scenario

**Feature**: 19:01 is inside overnight quiet

```
Active(19:01) == false for not-within-19h00m-to-06h30m
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
	req.At = "2026-08-24T19:01:00Z"
	return nil
}
```
