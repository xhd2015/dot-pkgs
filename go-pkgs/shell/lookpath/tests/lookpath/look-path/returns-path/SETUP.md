# Scenario

**Feature**: LookPath success returns the resolved path string

```
LookPathHit inject -> LookPath -> path string, err=nil
```

## Steps

1. Inject LookPath hit with an executable under WorkDir.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	hit := filepath.Join(req.WorkDir, "path", "mytool")
	writeExecutable(t, hit)
	req.LookPathHit = hit
	return nil
}
```
