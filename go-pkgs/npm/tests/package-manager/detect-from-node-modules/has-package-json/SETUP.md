# Scenario

**Feature**: package.json beside node_modules is detected

```
# HasPackageJSON via node_modules parent
package.json + node_modules -> HasPackageJSON true
```

## Steps

1. Write `package.json` and create empty `node_modules`.
2. Set `req.NodeModulesPath`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	projectDir := writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
	})
	req.NodeModulesPath = nodeModulesPath(t, projectDir)
	return nil
}```
