# Scenario

Worktree creation when destination is an existing directory.

mkdir existing-dir → [existing-dir]
mvd -w repo existing-dir → [(repo), (existing-dir/repo w:existing-dir/repo)]

## Steps
- Create a git repo at work/main.
- Create an existing destination directory.
- Set req.Args to run mvd -w into the existing directory.
- mvd should create the worktree inside the existing directory (not at it).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	srcRepo := filepath.Join(req.WorkRoot, "main")
	existingDir := filepath.Join(req.WorkRoot, "existing-dir")
	mkdirAll(t, srcRepo)
	initGitRepo(t, srcRepo)
	mkdirAll(t, existingDir)
	req.Args = []string{"-w", srcRepo, existingDir}
	return nil
}
```
