# Scenario

**Feature**: ModuleGoLine reads a module's `go` directive as major.minor

```
# go.mod go line -> go1.19 (patch dropped)
modDir go.mod -> ModuleGoLine -> go1.19 | error if missing
```

## Steps

1. Set `req.Op` to `goline`.
2. Create an isolated module dir (`t.TempDir()`).
3. Leaf `Setup` writes go.mod (or leaves the dir empty).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "goline"
	req.ModDir = t.TempDir()
	return nil
}
```
