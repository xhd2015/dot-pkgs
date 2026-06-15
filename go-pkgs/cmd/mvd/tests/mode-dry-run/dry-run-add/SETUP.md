## Steps
- Create a directory under WorkRoot.
- Run `mvd --dry-run --add dir` to dry-run adding it to history.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "myproject")
	mkdirAll(t, dir)
	req.Args = []string{"--dry-run", "--add", dir}
	return nil
}
```
