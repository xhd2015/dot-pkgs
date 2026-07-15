# Scenario

**Feature**: load of invalid JSON at the mirror entry path returns an error

```
# corrupt entry.json planted at expected mirror path
LoadCacheEntry -> error
```

## Steps

1. Compute expected mirror `entry.json` path for `RealPath` under `CacheRoot`.
2. Create intermediate directories and write non-JSON content (`not-valid-json{{{`).

```go
import (
	"os"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	path := expectedMirrorEntryPath(t, req.CacheRoot, req.RealPath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	writeFile(t, path, "not-valid-json{{{")
	return nil
}
```
