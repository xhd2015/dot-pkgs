## Steps
- Try to move by a basename "git-ops" that does not match any project.
- cwd is not the original project root.

```go
import (
	"os"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, dst)

	cwd := filepath.Join(req.WorkRoot, "cwd")
	mkdirAll(t, cwd)
	if err := os.Chdir(cwd); err != nil {
		return err
	}

	req.Args = []string{"git-ops", dst}
	return nil
}
```
