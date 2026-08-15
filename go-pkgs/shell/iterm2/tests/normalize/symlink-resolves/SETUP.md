# Scenario

**Feature**: symlink target directory resolves to canonical path in script

```go
import (
	"os"
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	real := filepath.Join(root, "real")
	if err := os.MkdirAll(real, 0755); err != nil {
		return err
	}
	link := filepath.Join(root, "link")
	if err := os.Symlink(real, link); err != nil {
		t.Skip("symlink not supported")
	}
	req.Dir = link
	return nil
}
```