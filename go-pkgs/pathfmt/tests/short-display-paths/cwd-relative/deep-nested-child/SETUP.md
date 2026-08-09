# Scenario

**Feature**: deeply nested child displays as `"a/b/c"`

```
# cwd rules (checked first)
strict child of cwd -> rel
```

## Steps

1. Set `req.Path` to `a/b/c` under project root.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Path = filepath.Join(req.Path, "a", "b", "c")
	return nil
}```
