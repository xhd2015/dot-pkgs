# Scenario

**Feature**: Scan returns root + nested modules, sorted by Dir, with correct paths

```
# root + app/ + nested/service/ go.mod under one git repo
root + app/go.mod + nested/service/go.mod -> scan.Scan -> [. , app, nested/service]
```

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd.
2. Add `app/go.mod` (`example.com/root/app`) and `nested/service/go.mod`
   (`example.com/root/service`).
3. Set `req.RootDir` (operation `scan` is set by the `basic/` grouping Setup).

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initBasicWorkspace(t, "example.com/root")

	writeModule(t, filepath.Join(ws, "app"), "example.com/root/app")
	writeModule(t, filepath.Join(ws, "nested", "service"), "example.com/root/service")

	req.RootDir = ws
	return nil
}
```
