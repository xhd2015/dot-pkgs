# Scenario

**Feature**: `LocalVersion` — run version command on bin via injectable runner

```
Bin + LookPath + RunVersion -> LocalVersion -> raw stdout | error
```

## Steps

1. Set `Operation=local-version`.
2. Leaves set runner success/fail fixtures.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "local-version"
	return nil
}
```
