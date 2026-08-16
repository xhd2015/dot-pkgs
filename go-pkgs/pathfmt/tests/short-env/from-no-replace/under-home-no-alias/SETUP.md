# Scenario

**Feature**: path under home with unrelated env aliases uses `~/...`, never `$HOME`

```
env has unrelated X=/other
path under home -> ~/...
```

## Steps

1. Inject an env alias whose value does not prefix the path.
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
	other := t.TempDir()
	req.Env = []string{envPair("X", other)}
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "short-env", "under-home-no-alias")
	return nil
}
```
