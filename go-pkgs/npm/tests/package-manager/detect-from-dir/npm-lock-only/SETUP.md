# Scenario

**Feature**: package-lock.json selects npm

```
# lockfile marker
package-lock.json -> Signal npm -> Manager npm
```

## Steps

1. Write `package.json` and `package-lock.json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json":      pkgJSONDemo,
		"package-lock.json": packageLockJSON,
	})
	return nil
}```
