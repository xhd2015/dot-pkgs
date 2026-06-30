# Scenario

**Feature**: wrk from main repo checkout

```
# cwd is the primary git checkout (not a linked worktree)
main repo checkout -> wrk -> {WRK_HOME}/worktrees/{basename}-{branch-token}-{YYYY-MM-DD}
```

## Steps

- Tests create a fresh git repo under `{WorkRoot}/myrepo` unless a leaf overrides setup.
- `req.RepoDir` points at the main checkout directory.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repoDir := filepath.Join(req.WorkRoot, "myrepo")
	req.RepoDir = repoDir
	return nil
}
```