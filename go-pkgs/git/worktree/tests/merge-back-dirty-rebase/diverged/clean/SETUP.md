# Scenario

**Feature**: diverged + clean + no --rm → direct rebase in source (existing behavior)

```
# worktree clean, Remove=false → rebase directly in source worktree
clean feat -> MergeBack(!Remove) -> rebase in source -> merge
```

## Steps

1. Source worktree is already clean from parent SETUP.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Remove = false
	req.MakeDirty = false
	return nil
}
```
