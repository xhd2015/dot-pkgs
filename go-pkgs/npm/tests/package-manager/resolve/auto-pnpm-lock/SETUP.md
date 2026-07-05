# Scenario

**Feature**: auto resolve picks pnpm from pnpm-lock.yaml

```
# auto detection + PATH
pnpm-lock.yaml -> candidate pnpm -> Manager pnpm
```

## Steps

1. Write `pnpm-lock.yaml` only.
2. Set `req.Pref` to `auto`.
3. Skip when `pnpm` is not on PATH.

```go
import (
	"testing"

	npm "github.com/xhd2015/dot-pkgs/go-pkgs/npm"
)

func Setup(t *testing.T, req *Request) error {
	requireManagerOnPath(t, npm.ManagerPnpm)
	req.ProjectDir = writeProject(t, map[string]string{
		"pnpm-lock.yaml": pnpmLockYAML,
	})
	req.Pref = "auto"
	return nil
}```
