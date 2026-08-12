# Scenario

**Feature**: path outside home stays absolute with no `~` prefix

```
# home rules
path not under home -> absolute unchanged
```

## Steps

1. Set `req.Path` to a path under the system temp directory (typically outside
   the user home). Create the directory so Abs is well-defined; no chdir.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	outside := filepath.Join(os.TempDir(), "doctest-tilde-home-outside-home")
	if err := os.MkdirAll(outside, 0o755); err != nil {
		return err
	}
	req.Path = outside
	return nil
}
```
