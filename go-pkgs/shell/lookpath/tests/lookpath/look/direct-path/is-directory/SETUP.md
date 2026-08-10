# Scenario

**Feature**: absolute directory path is skipped / rejected

```
absolute path that is a directory -> Look -> error
LookPath never called
```

## Steps

1. Create a directory at `$WorkDir/bin/mytool`.
2. Set `Name` to that absolute path.

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
	abs := filepath.Join(req.WorkDir, "bin", "mytool")
	if err := os.MkdirAll(abs, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	req.Name = abs
	return nil
}
```
