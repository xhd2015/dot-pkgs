# Scenario

**Feature**: Next returns the earliest valid fire >= from

```
Parse(expr); Next(anchor, from, UTC) -> (time, ok)
```

## Preconditions

- `req.Op` is `"next"`.
- Leaves set `req.Expr`, `req.Anchor`, `req.From` as RFC3339 UTC times.

## Steps

1. Set `req.Op` to `"next"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "next"
	return nil
}
```
