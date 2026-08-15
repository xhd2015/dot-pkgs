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
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	projectDir := writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
	})
	req.NodeModulesPath = nodeModulesPath(t, projectDir)
	return nil
}```
