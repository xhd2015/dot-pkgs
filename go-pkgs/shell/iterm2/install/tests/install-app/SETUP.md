# Scenario

**Feature**: `InstallApp` places extracted app at target (with optional backup)

```
InstallApp(extracted, target, Home) -> target tree; backup on replace
```

## Steps

1. Set `Operation=install-app`.
2. Leaves choose fresh / replace / default-home target.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "install-app"
	return nil
}
```
