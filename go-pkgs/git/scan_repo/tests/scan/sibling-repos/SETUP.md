# Scenario

**Feature**: two sibling repos discovered and sorted by path

```
Walk finds alpha/ and zebra/ -> two RepoTypeMain rows, path-sorted
```

## Steps

1. Create workspace with repos `zebra/` and `alpha/`.
2. Set `req.Roots` to the workspace.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	for _, name := range []string{"zebra", "alpha"} {
		dir := filepath.Join(root, name)
		mkdirAll(t, dir)
		fakeGitRepo(t, dir)
	}
	req.Roots = []string{root}
	return nil
}
```