# Scenario

**Feature**: file root (not directory) returns error

```
root is a regular file -> error: not a directory
```

## Steps

1. Create a regular file and use it as root.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	filePath := filepath.Join(root, "not-a-dir")
	writeFile(t, filePath, "content")
	req.Roots = []string{filePath}
	return nil
}
```