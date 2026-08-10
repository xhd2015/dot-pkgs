# Scenario

**Feature**: absolute executable path resolves with Via=direct

```
absolute path to 0755 file -> Look -> Path=abs, Via=direct
LookPath injectable never called
```

## Steps

1. Write executable at `$WorkDir/bin/mytool`.
2. Set `Name` to that absolute path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	abs := filepath.Join(req.WorkDir, "bin", "mytool")
	writeExecutable(t, abs)
	req.Name = abs
	return nil
}
```
