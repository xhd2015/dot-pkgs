# Scenario

**Feature**: `ShortEnv(path)` equals `ShortEnvFrom(path, os.Environ())`

```
path -> ShortEnv
     == ShortEnvFrom(path, os.Environ())
# no host-specific $VAR asserts
```

## Steps

1. Set a concrete absolute path under home (display form depends on live env).
2. Op is `current` from parent grouping.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	home := mustUserHome(t)
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "short-env", "current-wrapper")
	return nil
}
```
