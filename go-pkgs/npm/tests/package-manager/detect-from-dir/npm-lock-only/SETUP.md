# Scenario

**Feature**: package-lock.json selects npm

```
# lockfile marker
package-lock.json -> Signal npm -> Manager npm
```

## Steps

1. Write `package.json` and `package-lock.json`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json":      pkgJSONDemo,
		"package-lock.json": packageLockJSON,
	})
	return nil
}```
