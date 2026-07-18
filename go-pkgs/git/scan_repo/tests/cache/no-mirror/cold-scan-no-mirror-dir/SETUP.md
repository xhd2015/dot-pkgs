# Scenario

**Feature**: cold Scan with CacheRoot does not create `mirror/` directory

```
workspace/my-repo/.git + CacheRoot temp
  -> Scan(NoCache=false)
  -> Result discovers my-repo
  -> <CacheRoot>/mirror does not exist
  -> home/repos.json may exist (v2 index seed)
```

## Steps

1. Create workspace with one main repo `my-repo/`.
2. Set `req.Roots`; keep `NoCache=false` and parent temp `CacheRoot`.

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
	req.NoCache = false
	return nil
}

func fakeGitRepo(t *testing.T, dir string) {
	t.Helper()
	gitDir := filepath.Join(dir, ".git")
	mkdirAll(t, filepath.Join(gitDir, "objects"))
}
```
