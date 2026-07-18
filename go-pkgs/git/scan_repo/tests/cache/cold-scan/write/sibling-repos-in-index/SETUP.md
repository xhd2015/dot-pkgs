# Scenario

**Feature**: cold Scan indexes two sibling main checkouts

```
workspace/alpha + workspace/zebra
  -> Scan
  -> Result len 2 path-sorted; home/repos.json has both
```

## Steps

1. Create sibling mains `alpha/` and `zebra/`.
2. Set `req.Roots`.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	root := t.TempDir()
	for _, name := range []string{"alpha", "zebra"} {
		d := filepath.Join(root, name)
		mkdirAll(t, d)
		fakeGitRepo(t, d)
	}
	req.Roots = []string{root}
	return nil
}
```
