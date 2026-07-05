# Scenario

**Feature**: yarn.lock selects yarn

```
# lockfile marker
yarn.lock -> Signal yarn -> Manager yarn
```

## Steps

1. Write `package.json` and `yarn.lock`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
		"yarn.lock":    yarnLockStub,
	})
	return nil
}```
