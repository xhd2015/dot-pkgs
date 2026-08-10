# Scenario

**Feature**: directory path is not executable

```
directory -> IsExecutable -> false
```

## Steps

1. Create a directory under WorkDir; set `IsExecPath` to it.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	p := filepath.Join(req.WorkDir, "bindir")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	req.IsExecPath = p
	return nil
}
```
