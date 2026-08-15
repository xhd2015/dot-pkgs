# Scenario

**Feature**: diverged + dirty + no-rm → stash-based conflict detection

```
# Stash dirty changes from source, apply on tmp to test compatibility.
# Each sub-scenario exercises a different git conflict type during stash apply.
dirty feat -> stash push -> stash apply on tmp -> detect conflict or apply clean
```

## Steps

1. Each sub-scenario creates specific dirt on the source worktree.
2. The rebase target (main) makes a conflicting change.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Remove = false
	req.MakeDirty = true
	return nil
}
```
