# Scenario

**Feature**: Parse every-1h-at-4m sets Align to 4m

```
Parse("every-1h-at-4m") -> Align=4m
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
	return nil
}
```
