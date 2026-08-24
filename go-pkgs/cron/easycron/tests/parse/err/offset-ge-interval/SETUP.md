# Scenario

**Feature**: Parse rejects offset >= interval

```
Parse("every-1h-at-60m") -> error
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
	req.Expr = "every-1h-at-60m"
	return nil
}
```
