# Scenario

**Feature**: cold Scan writes non-repo intermediate directory entries

```
# workspace root is not a repo; my-repo is a child
workspace/ (non-repo) + workspace/my-repo/.git
  -> Scan
  -> LoadCacheEntry(cacheRoot, workspace): is_repo=false, children includes my-repo,
     scan_complete=true
```

## Steps

1. Create workspace with one main repo `my-repo/` under the scan root.
2. Set `req.Roots` to the workspace (the intermediate parent of the repo).

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
	// Extra empty dir so children is non-trivial if recorded.
	mkdirAll(t, filepath.Join(root, "notes"))
	req.Roots = []string{root}
	return nil
}
```
