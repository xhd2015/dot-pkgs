# Scenario

**Feature**: non-existent root path recorded as RootError; scan continues

```
root path does not exist -> RootErrors entry; err nil
```

## Steps

1. Set `req.Roots` to a path that does not exist under temp.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	req.Roots = []string{missing}
	return nil
}
```