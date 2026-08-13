# Scenario

**Feature**: a missing path is not an openable directory

```
dir=nonexistent path -> OpenConfig -> error
neither opener called
```

## Steps

1. Set `Dir` to a path under `WorkDir` that does not exist.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Dir = filepath.Join(req.WorkDir, "does-not-exist")
	return nil
}
```
