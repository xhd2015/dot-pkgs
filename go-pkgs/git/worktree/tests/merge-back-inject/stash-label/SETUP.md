# Scenario

**Feature**: inject StashLabel — dirty-diverged stash uses caller message

```
# MergeBackOptions.StashLabel set → stash push -m <label> during migrate
Caller sets StashLabel -> MergeBack dirty-diverged -> stash push -m label -> migrate
```

## Steps

1. Build diverged dirty fixture (helper).
2. Set a distinctive non-product `req.StashLabel`.
3. Leaf asserts success and that stash reflog retains the label.

## Context

- TmpDir may be set to WorkRoot-local parent so tests stay isolated without
  `WRK_HOME` (and without relying on default tmp placement).
- Product string `"wrk-merge-back"` must not be required for this path.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	setupDivergedDirty(t, req)

	// Isolate tmp worktrees under WorkRoot so the leaf focuses on StashLabel
	// without env or home-dir side effects.
	tmpParent := filepath.Join(req.WorkRoot, "inject-tmp-parent")
	if err := os.MkdirAll(tmpParent, 0755); err != nil {
		return err
	}
	req.TmpDir = tmpParent
	req.StashLabel = "doctest-inject-stash-label"
	return nil
}
```
