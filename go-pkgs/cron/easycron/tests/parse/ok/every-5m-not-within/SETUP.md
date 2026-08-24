# Scenario

**Feature**: Parse overnight not-within quiet window

```
Parse("every-5m-not-within-19h00m-to-06h30m") -> Quiet 19:00..06:30
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
	return nil
}
```
