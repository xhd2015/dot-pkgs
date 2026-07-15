# Scenario

**Feature**: cold Scan writes mirror entries for sibling main repos

```
# two sibling mains
workspace/zebra + workspace/alpha
  -> Scan
  -> both have is_repo cache entries; Result.Repos length 2, path-sorted
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
