# Scenario

**Feature**: custom TmpDir parent hosts the tmp rebase worktree

```
# Confirm plan.Commands[0].Dir is under inject TmpDir; flow succeeds and cleans up
TmpDir=inject-tmp-parent -> MergeBack -> rebase dir under TmpDir -> rebased-and-merged
  -> no leftover entries under TmpDir
```

## Preconditions

- Grouping set `req.TmpDir` to a dedicated empty directory under WorkRoot.
- Source is dirty diverged, Remove=false.

## Steps

1. Run MergeBack via root Run (Confirm captures rebase dir).
2. Assert success, dirt preserved, observed tmp path under inject parent, cleanup.

## Context

- Leaf of `tmp-dir/` — only custom parent path variant needed for P4 exit.

```go
import (
	"fmt"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	if req.SourcePath == "" || req.MainRepo == "" {
		return fmt.Errorf("custom-parent: ancestor fixture missing SourcePath/MainRepo")
	}
	if req.TmpDir == "" {
		return fmt.Errorf("custom-parent: req.TmpDir must be set by tmp-dir grouping")
	}
	if !hasDir(req.TmpDir) {
		return fmt.Errorf("custom-parent: inject TmpDir does not exist: %s", req.TmpDir)
	}
	return nil
}
```
