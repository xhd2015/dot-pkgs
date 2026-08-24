# Scenario

**Feature**: Parse maps expression strings to Expr or error

```
caller Expr string -> Parse -> Expr | error
```

## Preconditions

- `req.Op` is `"parse"`.
- Leaves set `req.Expr`.

## Steps

1. Set `req.Op` to `"parse"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "parse"
	return nil
}
```
