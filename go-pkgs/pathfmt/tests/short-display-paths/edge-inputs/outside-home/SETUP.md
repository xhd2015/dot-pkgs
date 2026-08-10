# Scenario

**Feature**: path outside home stays absolute when not under cwd

```
# fallback
otherwise -> absolute unchanged
```

## Steps

1. `chdir` to a temp cwd.
2. Set `req.Path` to another temp directory (system temp, typically outside home).

```go
import (
	"github.com/xhd2015/doctest/session"
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	chdirTo(t, t.TempDir())
	outside := filepath.Join(os.TempDir(), "doctest-display-outside-home")
	mkdirAll(t, outside)
	req.Path = outside
	return nil
}```
