# Scenario

**Feature**: Tokens splits on whitespace and drops empty fields

```
# Tokens
Tokens("  aid   user ") -> []string{"aid", "user"}
```

## Steps

1. Set `req.Op` to `"tokens"`, query `"  aid   user "`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "tokens"
	req.Query = "  aid   user "
	return nil
}
```
