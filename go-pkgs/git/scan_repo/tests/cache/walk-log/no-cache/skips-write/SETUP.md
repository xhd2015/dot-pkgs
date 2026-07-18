# Scenario

**Feature**: NoCache=true leaves home/ without walk.jsonl or walk.cursor.json

```
# one main under root, cache disabled for I/O
workspace/my-repo (fake .git)
  -> Scan(CacheRoot set, NoCache=true)
  -> Result still discovers my-repo
  -> <CacheRoot>/home/walk.jsonl absent
  -> <CacheRoot>/home/walk.cursor.json absent
```

## Steps

1. Create workspace with main `my-repo/`.
2. Set `req.Roots`; `NoCache=true` (grouping already set).

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	repo := filepath.Join(root, "my-repo")
	mkdirAll(t, repo)
	fakeGitRepo(t, repo)
	req.Roots = []string{root}
	req.NoCache = true
	req.ExpectWalkLog = false
	return nil
}
```
