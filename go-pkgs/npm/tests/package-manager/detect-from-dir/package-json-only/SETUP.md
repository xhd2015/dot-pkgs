# Scenario

**Feature**: package.json without indicators defaults to npm

```
# no indicators + package.json
package.json only -> default npm
```

## Steps

1. Write bare `package.json`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.ProjectDir = writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
	})
	return nil
}```
