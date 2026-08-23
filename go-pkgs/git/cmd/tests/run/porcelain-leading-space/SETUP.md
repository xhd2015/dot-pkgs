# Scenario

**Feature**: Run preserves leading spaces in `git status --porcelain` output

```
unstaged modify (XY=" M") -> Run(status --porcelain) -> output still starts with space
```

## Steps

1. Create temp repo with initial commit.
2. Modify a tracked file without staging.
3. Run `status --porcelain`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "proj")
	gitInitRepo(t, repoDir)
	gitInitialCommit(t, repoDir)
	writeFile(t, filepath.Join(repoDir, "README"), "dirty\n")
	req.Dir = repoDir
	req.Args = []string{"status", "--porcelain"}
	return nil
}
```
