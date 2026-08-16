# Scenario

**Feature**: a var whose value equals user home must not emit `$THAT`; use `~`

```
MYHOME=<user home>
path under home -> ~/...  never $MYHOME/...
```

## Steps

1. Inject `MYHOME` equal to user home (eligible name, home-valued).
2. Path under home.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	home := mustUserHome(t)
	req.Env = []string{envPair("MYHOME", home)}
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "short-env", "home-valued-var")
	return nil
}
```
