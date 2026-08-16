# Scenario

**Feature**: path under an env alias displays as `$NAME` + remainder

```
X=/tmp/proj
path=/tmp/proj/pkg/a
-> $X/pkg/a
```

## Steps

1. Create temp project root; set env `X=<root>`; path to `root/pkg/a`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	proj := t.TempDir()
	req.Env = []string{envPair("X", proj)}
	req.Path = filepath.Join(proj, "pkg", "a")
	return nil
}
```
