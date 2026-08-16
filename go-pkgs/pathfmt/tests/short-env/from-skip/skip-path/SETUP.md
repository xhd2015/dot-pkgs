# Scenario

**Feature**: `PATH`-style multi-directory values are not eligible aliases

```
PATH=/dirA:/dirB
path under /dirA -> not $PATH/...  (TildeHome or abs)
```

## Steps

1. Create two temp dirs; set `PATH=dirA:dirB`.
2. Path under dirA.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	dirA := t.TempDir()
	dirB := t.TempDir()
	req.Env = []string{envPair("PATH", dirA+string(filepath.ListSeparator)+dirB)}
	req.Path = filepath.Join(dirA, "bin", "tool")
	return nil
}
```
