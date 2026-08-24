# Scenario

**Feature**: Parse rejects non-every prefix

```
Parse("daily-1h") -> error
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
	req.Expr = "daily-1h"
	return nil
}
```
