# Scenario

**Feature**: missing image path returns a not-found error

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ImagePath = filepath.Join(t.TempDir(), "missing.png")
	return nil
}
```
