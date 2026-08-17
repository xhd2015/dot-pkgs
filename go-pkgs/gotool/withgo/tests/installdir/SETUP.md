# Scenario

**Feature**: DefaultInstallDir is `$HOME/installed` from UserHomeDir

```
# read-only; no env mutation
DefaultInstallDir -> filepath.Join(UserHomeDir(), "installed")
```

## Steps

1. Set `req.Op` to `installdir`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "installdir"
	return nil
}
```
