# Scenario

**Feature**: `DetectFromNodeModules` inspects parent project via node_modules path

```
# node_modules path + optional .pnpm store
node_modules path -> DetectFromNodeModules -> Trace.Manager + HasPackageJSON
```

## Steps

1. Leaf `Setup` writes temp project files and sets `req.NodeModulesPath`.
2. `req.Op` is `detect-node-modules`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Op = "detect-node-modules"
	return nil
}
```