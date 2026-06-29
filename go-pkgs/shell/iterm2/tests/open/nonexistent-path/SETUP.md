# Scenario

**Feature**: stat error for missing directory

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Dir = filepath.Join(t.TempDir(), "does-not-exist")
	return nil
}
```