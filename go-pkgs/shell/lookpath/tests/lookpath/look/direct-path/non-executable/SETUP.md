# Scenario

**Feature**: absolute non-executable file is rejected (Via not success)

```
absolute path to 0644 file -> Look -> error
LookPath never called
```

## Steps

1. Write a non-executable regular file at `$WorkDir/bin/mytool`.
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
	writeNonExecutable(t, abs)
	req.Name = abs
	return nil
}
```
