# Scenario

**Feature**: `VerifyScriptable` with injectable runner (no real osascript)

```
VerifyScriptable(Runner) -> version or error
```

## Steps

1. Set `Operation=verify-scriptable`.
2. Leaves set success version or failure flag.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "verify-scriptable"
	return nil
}
```
