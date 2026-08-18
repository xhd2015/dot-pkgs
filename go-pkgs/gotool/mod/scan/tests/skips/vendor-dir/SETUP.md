# Scenario

**Feature**: name skip — a `vendor/` directory is pruned (existing kool behavior, sanity)

```
# vendor dirs hold third-party deps, never first-class modules of this workspace
root + vendor/p/go.mod -> scan.Scan -> [.]  (vendor subtree absent)
```

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd.
2. Add `vendor/p/go.mod`.
3. Set `req.RootDir` (operation `scan` is set by the `skips/` grouping Setup).

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initSkipRoot(t, "example.com/root")
	writeModule(t, filepath.Join(ws, "vendor", "p"), "example.com/root/vendor-p")
	req.RootDir = ws
	return nil
}
```
