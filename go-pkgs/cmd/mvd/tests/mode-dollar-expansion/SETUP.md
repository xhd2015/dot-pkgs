## Steps
- Set HOME and X environment variables for lls dollar-expansion support.
- HOME points to `req.WorkRoot/.lls-home` where the lls config with `"envs":["X"]` resides.
- X resolves to `req.WorkRoot/projects`.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	homeDir := filepath.Join(req.WorkRoot, ".lls-home")
	t.Setenv("HOME", homeDir)
	t.Setenv("X", filepath.Join(req.WorkRoot, "projects"))
	return nil
}
```
