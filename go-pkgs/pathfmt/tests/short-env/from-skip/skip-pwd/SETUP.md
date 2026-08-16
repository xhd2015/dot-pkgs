# Scenario

**Feature**: `PWD` is never used as a `$PWD` display alias

```
PWD=/proj
path=/proj/src
-> not $PWD/src
```

## Steps

1. Create temp project; set `PWD=<proj>`.
2. Path is a child of proj.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	proj := t.TempDir()
	req.Env = []string{envPair("PWD", proj)}
	req.Path = filepath.Join(proj, "src")
	return nil
}
```
