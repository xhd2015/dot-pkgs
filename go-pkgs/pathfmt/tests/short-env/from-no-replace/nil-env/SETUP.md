# Scenario

**Feature**: nil env is treated like empty — `TildeHome` only, no `os.Environ` magic

```
env=nil
path under home -> ~/...
```

## Steps

1. Leave `req.Env` as `nil`.
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
	req.Env = nil
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "short-env", "nil-env")
	return nil
}
```
