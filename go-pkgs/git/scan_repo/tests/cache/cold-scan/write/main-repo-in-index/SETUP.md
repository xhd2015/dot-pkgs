# Scenario

**Feature**: cold Scan seeds home/repos.json for a single main checkout

```
workspace/my-repo/.git (dir)
  -> Scan(CacheRoot, NoCache=false)
  -> LoadRepoIndex(home) contains my-repo main
  -> no mirror/
```

## Steps

1. Create workspace with one fake main repo at `my-repo/`.
2. Set `req.Roots` to the workspace.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "my-repo")
	mkdirAll(t, repoDir)
	fakeGitRepo(t, repoDir)
	req.Roots = []string{root}
	return nil
}
```
