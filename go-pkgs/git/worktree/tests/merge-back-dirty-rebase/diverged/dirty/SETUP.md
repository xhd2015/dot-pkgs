# Scenario

**Feature**: diverged + dirty + no --rm → tmp worktree path

```
# worktree dirty, Remove=false → rebase in configured temporary-worktree parent
dirty feat -> MergeBack(!Remove) -> create tmp worktree -> rebase there -> merge -> cleanup
```

## Steps

1. Make feature worktree dirty.
2. Each sub-scenario exercises a different outcome of the tmp-worktree flow.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	makeDirty(t, req.SourcePath)
	req.Remove = false
	req.MakeDirty = true
	return nil
}
```
