# Scenario

**Feature**: <target-dir> absent but parent exists → spawn worktree exactly at <target-dir>

```
# parent {WorkRoot} exists, <target-dir> {WorkRoot}/wt does not
myrepo (main) -> wrk myrepo {WorkRoot}/wt -> worktree at {WorkRoot}/wt (no naming suffix on path)
# branch still defaults to main-{date}; WRK_HOME ignored
```

## Steps

1. Source repo `myrepo` on `main` is initialized by the parent setup.
2. Set `req.SpawnDir = {WorkRoot}/wt` (does not exist; parent `{WorkRoot}` exists).
3. Run `wrk myrepo {WorkRoot}/wt` from process cwd `{WorkRoot}`.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	return nil
}
```
