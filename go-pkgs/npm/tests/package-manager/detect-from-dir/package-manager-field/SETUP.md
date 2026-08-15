# Scenario

**Feature**: packageManager field alone selects pnpm

```
# package.json packageManager field
packageManager pnpm@11 -> Signal pnpm -> Manager pnpm
```

## Steps

1. Write `package.json` with `packageManager` pnpm only.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json": pkgJSONPnpmPM,
	})
	return nil
}```
