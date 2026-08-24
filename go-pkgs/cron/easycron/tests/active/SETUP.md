# Scenario

**Feature**: Active reports quiet-window membership

```
Parse(expr); Active(at, UTC) -> bool
```

## Preconditions

- `req.Op` is `"active"`.
- Leaves set `req.Expr` and `req.At`.

## Steps

1. Set `req.Op` to `"active"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "active"
	return nil
}
```
