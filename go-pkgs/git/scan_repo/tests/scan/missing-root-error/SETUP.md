# Scenario

**Feature**: non-existent root path returns error

```
root path does not exist -> error wrapping path
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