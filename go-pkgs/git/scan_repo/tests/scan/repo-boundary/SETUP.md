# Scenario

**Feature**: nested repo inside discovered repo is skipped (SkipDir boundary)

```
outer/.git found -> do not descend into outer/ -> inner/.git not reported
```

## Steps

1. Create `outer/` with `.git` and nested `inner/` with its own `.git`.
2. Set `req.Roots` to workspace.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	outer := filepath.Join(root, "outer")
	inner := filepath.Join(outer, "inner")
	mkdirAll(t, inner)
	fakeGitRepo(t, outer)
	fakeGitRepo(t, inner)
	req.Roots = []string{root}
	return nil
}
```