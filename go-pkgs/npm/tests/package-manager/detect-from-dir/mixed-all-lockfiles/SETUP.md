# Scenario

**Feature**: pnpm wins when all lockfiles and packageManager npm are present

```
# priority pnpm > bun > npm > yarn
all lockfiles + packageManager npm -> Manager pnpm
```

## Steps

1. Write every lockfile plus `packageManager` npm field.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json":       pkgJSONNpmPM,
		"pnpm-lock.yaml":     pnpmLockYAML,
		"bun.lock":           bunLockJSON,
		"bun.lockb":          bunLockbStub,
		"package-lock.json":  packageLockJSON,
		"yarn.lock":          yarnLockStub,
	})
	return nil
}```
