# Scenario

**Feature**: repo beyond MaxDepth is excluded

```
root/level1/level2/deep-repo/.git at depth 3 -> MaxDepth=2 excludes it
```

## Steps

1. Create nested layout `level1/level2/deep-repo/` with `.git` at depth 3 from root.
2. Set `req.MaxDepth` to 2.
3. Set `req.Roots` to workspace.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	deep := filepath.Join(root, "level1", "level2", "deep-repo")
	mkdirAll(t, deep)
	fakeGitRepo(t, deep)
	req.Roots = []string{root}
	req.MaxDepth = 2
	return nil
}
```