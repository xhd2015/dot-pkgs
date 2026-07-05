# Scenario

**Feature**: auto resolve prefers pnpm when all lockfiles exist

```
# priority + PATH
all lockfiles -> candidate pnpm first -> Manager pnpm
```

## Steps

1. Write all lockfiles and `packageManager` npm field.
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
		"package.json":       pkgJSONNpmPM,
		"pnpm-lock.yaml":     pnpmLockYAML,
		"bun.lock":           bunLockJSON,
		"bun.lockb":          bunLockbStub,
		"package-lock.json":  packageLockJSON,
		"yarn.lock":          yarnLockStub,
	})
	req.Pref = "auto"
	return nil
}```
