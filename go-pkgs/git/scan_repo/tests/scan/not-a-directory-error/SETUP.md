# Scenario

**Feature**: file root (not directory) recorded as RootError; scan continues

```
root is a regular file -> RootErrors entry; err nil
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