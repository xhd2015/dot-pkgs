# Scenario

**Feature**: bun.lockb selects bun

```
# lockfile marker
bun.lockb -> Signal bun -> Manager bun
```

## Steps

1. Write `package.json` and `bun.lockb`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
		"bun.lockb":    bunLockbStub,
	})
	return nil
}```
