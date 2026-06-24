# Scenario

**Feature**: single main repo discovered under root

```
Walk finds .git directory -> RepoTypeMain row
```

## Steps

1. Create workspace with one repo at `my-repo/`.
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