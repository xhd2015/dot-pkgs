# Scenario

**Feature**: load when no entry file exists returns ok=false without error

```
# missing entry.json
LoadCacheEntry -> (zero, false, nil)
```

## Steps

1. Set a distinct `RealPath` that was never saved.
2. Assert preconditions: the expected mirror `entry.json` path does not exist
   (fresh `CacheRoot` from parent).

```go
import (
	"fmt"
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.RealPath = "/Users/xhd2015/Projects/never-written"
	path := expectedMirrorEntryPath(t, req.CacheRoot, req.RealPath)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("precondition failed: entry already exists at %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat mirror path: %w", err)
	}
	return nil
}
```
