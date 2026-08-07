# Scenario

**Feature**: `ExtractApp` unpacks zip and locates `iTerm.app`

```
ExtractApp(zip, workDir) -> appPath under workDir or error
```

## Steps

1. Set `Operation=extract`.
2. Leaves control whether zip contains `iTerm.app`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "extract"
	return nil
}
```
