# Scenario

**Feature**: the name `HOME` is never emitted as `$HOME` (home uses `~`)

```
HOME=<user home>
path under home -> ~/...  never $HOME/...
```

## Steps

1. Inject `HOME` equal to `os.UserHomeDir()`.
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
	req.Env = []string{envPair("HOME", home)}
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "short-env", "skip-home-name")
	return nil
}
```
