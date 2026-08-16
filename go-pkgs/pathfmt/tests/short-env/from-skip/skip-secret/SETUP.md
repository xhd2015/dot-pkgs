# Scenario

**Feature**: secret-ish names containing `KEY` are not eligible aliases

```
FOO_KEY=/secret/dir
path=/secret/dir/a
-> not $FOO_KEY/a
```

## Steps

1. Create temp dir; set `FOO_KEY=<dir>`.
2. Path is a child of that dir.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	secret := t.TempDir()
	req.Env = []string{envPair("FOO_KEY", secret)}
	req.Path = filepath.Join(secret, "a")
	return nil
}
```
