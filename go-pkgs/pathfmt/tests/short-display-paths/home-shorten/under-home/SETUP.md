# Scenario

**Feature**: cache-like path under home displays as `~/...`

```
# home shorten (when not under cwd)
path under home -> "~" + suffix
```

## Steps

1. Set `req.Path` to a doctest mapping-gen cache-like path under home.

```go
import (
	"github.com/xhd2015/doctest/session"
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	req.Path = filepath.Join(home, "Library", "Caches", "doctest", "mapping-gen", "example", "proj")
	return nil
}```
