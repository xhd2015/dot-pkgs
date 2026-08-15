# Scenario

**Feature**: `NoCache=true` discovers repos but writes no cache artifacts under CacheRoot

```
workspace/my-repo + CacheRoot
  -> Scan(NoCache=true)
  -> Result has my-repo
  -> CacheRoot has no home/repos.json, no walk.jsonl, no mirror/
```

## Steps

1. Create workspace with one main repo.
2. Keep NoCache=true from parent.

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
	req.NoCache = true
	return nil
}
```
