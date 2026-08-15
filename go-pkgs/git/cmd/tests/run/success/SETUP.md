# Scenario

**Feature**: Run succeeds inside a git work tree

```
git init + commit -> Run(rev-parse --is-inside-work-tree) -> "true"
```

## Steps

1. Create temp repo with initial commit.
2. Run `rev-parse --is-inside-work-tree`.

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
	req.Dir = repoDir
	req.Args = []string{"rev-parse", "--is-inside-work-tree"}
	return nil
}
```
