# Scenario

**Feature**: reject non-directory paths

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	writeFile(t, file, "x")
	req.Dir = file
	return nil
}
```