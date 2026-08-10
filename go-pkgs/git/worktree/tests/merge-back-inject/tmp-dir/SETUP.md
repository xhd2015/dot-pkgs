# Scenario

**Feature**: inject TmpDir — tmp worktree parent is caller-supplied

```
# MergeBackOptions.TmpDir set → create tmp worktree under that parent
Caller sets TmpDir -> MergeBack dirty-diverged -> rebase -C <TmpDir/.../tmp-rebase>
```

## Steps

1. Build diverged dirty fixture (helper).
2. Allocate a custom parent under WorkRoot and set `req.TmpDir`.
3. Leaves assert the Confirm plan's rebase dir is under that parent.

## Context

- StashLabel left empty (library default); not under test on this branch.
- No `WRK_HOME` env; path comes only from options.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	setupDivergedDirty(t, req)

	tmpParent := filepath.Join(req.WorkRoot, "inject-tmp-parent")
	if err := os.MkdirAll(tmpParent, 0755); err != nil {
		return err
	}
	req.TmpDir = tmpParent
	req.StashLabel = ""
	return nil
}
```
