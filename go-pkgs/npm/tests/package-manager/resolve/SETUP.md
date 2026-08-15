# Scenario

**Feature**: `Resolve` picks explicit or auto-detected manager with PATH checks

```
# explicit pref or auto detection
projectDir + pref -> Resolve -> Manager or error
```

## Steps

1. Leaf `Setup` writes fixtures, sets `req.ProjectDir` and `req.Pref`.
2. `req.Op` is `resolve`.
3. Leaves requiring a specific CLI call `requireManagerOnPath` before running.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "resolve"
	return nil
}
```