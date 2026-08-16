# Scenario

**Feature**: empty env slice falls back to `TildeHome` only

```
env=[]
path under home -> ~/...
```

## Steps

1. Set `req.Env` to a non-nil empty slice.
2. Set path under user home.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	home := mustUserHome(t)
	req.Env = []string{}
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "short-env", "empty-env")
	return nil
}
```
