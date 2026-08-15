# Scenario

**Feature**: stat error for missing directory

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = filepath.Join(t.TempDir(), "does-not-exist")
	return nil
}
```