# Scenario

**Feature**: pnpm-lock.yaml alone selects pnpm

```
# lockfile marker
pnpm-lock.yaml -> Signal pnpm -> Manager pnpm
```

## Steps

1. Write `package.json` and `pnpm-lock.yaml`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json":   pkgJSONDemo,
		"pnpm-lock.yaml": pnpmLockYAML,
	})
	return nil
}```
