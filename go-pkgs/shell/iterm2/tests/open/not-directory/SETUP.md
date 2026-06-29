# Scenario

**Feature**: reject non-directory paths

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	writeFile(t, file, "x")
	req.Dir = file
	return nil
}
```