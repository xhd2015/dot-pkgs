# Scenario

**Feature**: bun.lock selects bun

```
# lockfile marker
bun.lock -> Signal bun -> Manager bun
```

## Steps

1. Write `package.json` and `bun.lock`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
		"bun.lock":     bunLockJSON,
	})
	return nil
}```
