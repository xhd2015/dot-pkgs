## Steps
- Create a non-git directory.
- Try to use -w on it.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	nonGit := filepath.Join(req.WorkRoot, "not-a-repo")
	mkdirAll(t, nonGit)

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", nonGit, wtDir}
	return nil
}
```
