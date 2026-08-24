# Scenario

**Feature**: Parse rejects not-within without -to-

```
Parse("every-5m-not-within-19h00m") -> error
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
	req.Expr = "every-5m-not-within-19h00m"
	return nil
}
```
