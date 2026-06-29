# Scenario

**Feature**: `Options.IgnoreDirBasenames` adds custom basename ignores anywhere in the tree

```
caller IgnoreDirBasenames=[scratch] -> Walk skips scratch -> hidden-repo not discovered
```

## Steps

1. Create `scratch/hidden-repo/` with fake `.git`.
2. Set `req.IgnoreDirBasenames` to `scratch`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	hidden := filepath.Join(root, "scratch", "hidden-repo")
	mkdirAll(t, hidden)
	fakeGitRepo(t, hidden)
	req.Roots = []string{root}
	req.IgnoreDirBasenames = []string{"scratch"}
	return nil
}
```