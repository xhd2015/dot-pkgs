# Scenario

**Feature**: cold Scan writes a mirror entry for a single main checkout

```
# one main under root
workspace/my-repo/.git (dir)
  -> Scan(CacheRoot, NoCache=false)
  -> LoadCacheEntry(cacheRoot, my-repo) is_repo=true, repo_type=main
```

## Steps

1. Create workspace with one fake main repo at `my-repo/`.
2. Set `req.Roots` to the workspace.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	repoDir := filepath.Join(root, "my-repo")
	mkdirAll(t, repoDir)
	fakeGitRepo(t, repoDir)
	req.Roots = []string{root}
	return nil
}
```
