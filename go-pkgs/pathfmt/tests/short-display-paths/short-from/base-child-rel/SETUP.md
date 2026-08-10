# Scenario

**Feature**: path under non-home baseDir displays as relative child

```
baseDir = temp/proj
path    = temp/proj/child
ShortFrom -> "child"
```

## Steps

1. Create temp project with `child/`.
2. Set BaseDir to project root; Path to child.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	proj := t.TempDir()
	child := filepath.Join(proj, "child")
	mkdirAll(t, child)
	req.BaseDir = proj
	req.Path = child
	return nil
}
```
