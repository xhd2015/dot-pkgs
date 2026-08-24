# Scenario

**Feature**: Parse every-1h yields 1h interval with no modifiers

```
Parse("every-1h") -> Interval=1h
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
	return nil
}
```
