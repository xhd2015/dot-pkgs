# Scenario

**Feature**: git init with no remotes yields empty Remotes slice

```
git init, no remote add -> Remotes empty
```

## Steps

1. Create real git repo without remotes.
2. Set `req.Roots` to parent of repo.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	repoDir := filepath.Join(root, "solo")
	gitInitRepo(t, repoDir)
	req.Roots = []string{root}
	return nil
}
```