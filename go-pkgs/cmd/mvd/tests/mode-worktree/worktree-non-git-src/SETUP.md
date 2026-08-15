# Scenario

Error when source is not a git repo.

mvd -w plain-dir wt → error → not a git repository

## Steps
- Create a non-git directory.
- Try to use -w on it.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	nonGit := filepath.Join(req.WorkRoot, "not-a-repo")
	mkdirAll(t, nonGit)

	wtDir := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"-w", nonGit, wtDir}
	return nil
}
```
