# Scenario

**Feature**: yarn.lock selects yarn

```
# lockfile marker
yarn.lock -> Signal yarn -> Manager yarn
```

## Steps

1. Write `package.json` and `yarn.lock`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
		"yarn.lock":    yarnLockStub,
	})
	return nil
}```
