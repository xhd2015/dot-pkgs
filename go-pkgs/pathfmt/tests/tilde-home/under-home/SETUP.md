# Scenario

**Feature**: absolute path under home displays as `"~/..."` with native separators

```
# home rules
path under home -> "~" + strings.TrimPrefix(abs, home)
```

## Steps

1. Set `req.Path` to a distinctive absolute path under the user home
   (need not exist on disk; Abs + string prefix only).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	home := mustUserHome(t)
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "tilde-home", "marker-under-home")
	return nil
}
```
