# Scenario

**Feature**: `NoCache=true` performs discovery but writes no mirror files

```
# NoCache skip
workspace/my-repo + CacheRoot set + NoCache=true
  -> Scan discovers my-repo
  -> CacheRoot/mirror has no entry.json (mirror absent or empty of entries)
```

## Steps

1. Create workspace with one main repo `my-repo/`.
2. Set `req.Roots`; parent already set `NoCache=true` and temp `CacheRoot`.

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
	req.NoCache = true
	return nil
}
```
