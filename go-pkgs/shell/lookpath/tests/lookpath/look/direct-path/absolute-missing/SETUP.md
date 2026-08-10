# Scenario

**Feature**: missing absolute path errors without consulting later stages

```
absolute missing path -> Look -> error
LookPath never called (no fallthrough)
```

## Steps

1. Set `Name` to a non-existent absolute path under WorkDir.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Name = filepath.Join(req.WorkDir, "missing", "mytool")
	return nil
}
```
