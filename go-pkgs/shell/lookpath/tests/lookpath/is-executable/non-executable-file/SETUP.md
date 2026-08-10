# Scenario

**Feature**: regular file mode 0644 is not executable

```
0644 file -> IsExecutable -> false
```

## Steps

1. Write non-executable under WorkDir; set `IsExecPath`.

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
	writeNonExecutable(t, p)
	req.IsExecPath = p
	return nil
}
```
