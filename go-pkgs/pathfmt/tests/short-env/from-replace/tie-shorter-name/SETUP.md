# Scenario

**Feature**: when two aliases share the same path value, the shorter name wins

```
X=/tmp/proj
PROJECT_X=/tmp/proj
path=/tmp/proj/src
-> $X/src  (not $PROJECT_X/src)
```

## Steps

1. Create temp dir; set both `X` and `PROJECT_X` to that dir.
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
		envPair("PROJECT_X", proj),
		envPair("X", proj),
	}
	req.Path = filepath.Join(proj, "src")
	return nil
}
```
