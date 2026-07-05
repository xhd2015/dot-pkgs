# Scenario

**Feature**: node_modules/.pnpm store layout signals pnpm

```
# pnpm content-addressable store
node_modules/.pnpm -> Signal pnpm -> Manager pnpm
```

## Steps

1. Write `package.json` and create `node_modules/.pnpm/pkg@1.0.0`.
2. Set `req.NodeModulesPath` to the project `node_modules` directory.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	projectDir := writeProject(t, map[string]string{
		"package.json": pkgJSONDemo,
	})
	mkdirProjectDir(t, projectDir, filepath.Join("node_modules", ".pnpm", "pkg@1.0.0"))
	req.NodeModulesPath = nodeModulesPath(t, projectDir)
	return nil
}```
