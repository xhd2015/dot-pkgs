# Scenario

**Feature**: bun.lock selects bun

```
# lockfile marker
bun.lock -> Signal bun -> Manager bun
```

## Steps

1. Write `package.json` and `bun.lock`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
		"bun.lock":     bunLockJSON,
	})
	return nil
}```
