# Scenario

**Feature**: pnpm-lock.yaml alone selects pnpm

```
# lockfile marker
pnpm-lock.yaml -> Signal pnpm -> Manager pnpm
```

## Steps

1. Write `package.json` and `pnpm-lock.yaml`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json":   pkgJSONDemo,
		"pnpm-lock.yaml": pnpmLockYAML,
	})
	return nil
}```
