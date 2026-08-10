# Scenario

**Feature**: regular file mode 0755 is executable

```
0755 file -> IsExecutable -> true
```

## Steps

1. Write executable under WorkDir; set `IsExecPath`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	p := filepath.Join(req.WorkDir, "bin", "tool")
	writeExecutable(t, p)
	req.IsExecPath = p
	return nil
}
```
