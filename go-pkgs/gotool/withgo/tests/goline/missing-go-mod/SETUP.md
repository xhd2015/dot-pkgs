# Scenario

**Feature**: a directory with no go.mod is an error

```
# empty module dir
modDir without go.mod -> ModuleGoLine -> error
```

## Steps

1. Point `req.ModDir` at an empty subdirectory (no go.mod).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	empty := filepath.Join(req.ModDir, "empty")
	if err := os.MkdirAll(empty, 0755); err != nil {
		return err
	}
	req.ModDir = empty
	return nil
}
```
