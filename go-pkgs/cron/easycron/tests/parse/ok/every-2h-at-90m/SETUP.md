# Scenario

**Feature**: Parse every-2h-at-90m accepts offset under 2h

```
Parse("every-2h-at-90m") -> Align=90m
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
	req.Expr = "every-2h-at-90m"
	return nil
}
```
