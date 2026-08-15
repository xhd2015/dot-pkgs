# Scenario

**Feature**: bun beats npm when both lockfiles exist

```
# priority bun > npm
bun.lock + package-lock.json -> Manager bun
```

## Steps

1. Write `bun.lock` and `package-lock.json` without package.json.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"bun.lock":          bunLockJSON,
		"package-lock.json": packageLockJSON,
	})
	return nil
}```
