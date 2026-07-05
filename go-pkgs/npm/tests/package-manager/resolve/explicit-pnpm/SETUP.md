# Scenario

**Feature**: explicit pnpm pref resolves when CLI is on PATH

```
# explicit pref bypasses detection
pref pnpm + PATH -> Manager pnpm
```

## Steps

1. Create empty project directory.
2. Set `req.Pref` to `pnpm`.
3. Skip when `pnpm` is not on PATH.

```go
import (
	"testing"

	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

func Setup(t *testing.T, req *Request) error {
	requireManagerOnPath(t, npm.ManagerPnpm)
	req.ProjectDir = writeProject(t, nil)
	req.Pref = "pnpm"
	return nil
}```
