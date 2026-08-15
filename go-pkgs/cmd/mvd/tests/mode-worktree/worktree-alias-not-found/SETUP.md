# Scenario

Error when alias is not registered for -w.

mvd -w NOSUCHALIAS wt → error → not found → "not a git repository"

## Steps
- Change CWD to an empty directory to ensure no local files shadow the alias name.
- Run mvd -w with an unregistered alias name.

```go
import (
	"os"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	cwd := filepath.Join(req.WorkRoot, "cwd")
	mkdirAll(t, cwd)
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", "NOSUCHALIAS", wtDir}
	return nil
}
```
