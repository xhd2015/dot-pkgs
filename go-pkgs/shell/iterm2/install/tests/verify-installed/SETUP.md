# Scenario

**Feature**: `VerifyInstalled` checks bundle layout + BundleID

```
VerifyInstalled(appPath) -> nil or error
```

## Steps

1. Set `Operation=verify-installed`.
2. Leaves build mini-bundle variants under WorkDir.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Operation = "verify-installed"
	return nil
}
```
