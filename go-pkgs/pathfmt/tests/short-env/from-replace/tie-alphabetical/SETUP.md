# Scenario

**Feature**: same path value and same name length → alphabetical name wins

```
BB=/tmp/proj
AA=/tmp/proj
path=/tmp/proj/file
-> $AA/file  (not $BB/file)
```

## Steps

1. Create temp dir; set `AA` and `BB` (equal length) to that dir.
2. Path is a child of the dir.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	proj := t.TempDir()
	req.Env = []string{
		envPair("BB", proj),
		envPair("AA", proj),
	}
	req.Path = filepath.Join(proj, "file")
	return nil
}
```
