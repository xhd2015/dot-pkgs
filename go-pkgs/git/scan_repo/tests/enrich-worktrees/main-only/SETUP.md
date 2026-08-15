# Scenario

**Feature**: single main repo lists itself as the only worktree

```
main repo only -> Worktrees[{main, IsMain=true}] on main row
```

## Steps

1. Init repo with initial commit (required for worktree list).
2. Set `req.Roots` to parent of repo.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if !gitAvailable(t) {
		return nil
	}
	root := t.TempDir()
	mainDir := filepath.Join(root, "main")
	gitInitRepo(t, mainDir)
	gitInitialCommit(t, mainDir)
	req.Roots = []string{root}
	return nil
}
```