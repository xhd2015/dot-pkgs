# Scenario

**Feature**: load when no repos.json exists returns ok=false without error

```
# missing repos.json
LoadRepoIndex -> (zero/empty index, false, nil)
```

## Steps

1. Set `Universe` to `"home"`.
2. Assert precondition: expected `home/repos.json` path does not exist under
   fresh `CacheRoot`.

```go
import (
	"fmt"
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Universe = "home"
	path := expectedRepoIndexPath(t, req.CacheRoot, req.Universe)
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("precondition failed: repos.json already exists at %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat index path: %w", err)
	}
	return nil
}
```
