# Scenario

--dry-run -w with non-git src: validation error still fires.

mvd --dry-run -w plain-dir wt → error → not a git repository

## Steps
- Create a non-git directory, then try `mvd --dry-run -w nosrc dst`.
- Validation should still fail because SRC is not a git repo.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	nonGit := filepath.Join(req.WorkRoot, "not-a-repo")
	mkdirAll(t, nonGit)
	dst := filepath.Join(req.WorkRoot, "feature")
	req.Args = []string{"--dry-run", "-w", nonGit, dst}
	return nil
}
```
