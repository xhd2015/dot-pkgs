# Scenario

**Feature**: ErrNotInstalled when injectable Installed is false

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = t.TempDir()
	req.UseInstalledOverride = true
	req.InstalledOK = false
	return nil
}
```