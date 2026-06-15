## Steps
- Create a non-git directory, then try `mvd --dry-run -w nosrc dst`.
- Validation should still fail because SRC is not a git repo.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	nonGit := filepath.Join(req.WorkRoot, "not-a-repo")
	mkdirAll(t, nonGit)
	dst := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"--dry-run", "-w", nonGit, dst}
	return nil
}
```
