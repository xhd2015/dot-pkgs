# Scenario

**Feature**: empty baseDir falls back to process cwd (same as Short)

```
baseDir = ""
cwd     = temp/proj
path    = temp/proj/child
ShortFrom -> "child"
```

## Steps

1. Save/restore cwd; chdir to temp project with `child/`.
2. Set BaseDir empty; Path to child absolute.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	saveAndRestoreCwd(t)
	proj := t.TempDir()
	child := filepath.Join(proj, "child")
	mkdirAll(t, child)
	chdirTo(t, proj)
	req.BaseDir = ""
	req.Path = child
	return nil
}
```
