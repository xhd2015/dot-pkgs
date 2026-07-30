# Scenario

**Feature**: List includes main + linked; ListLinked excludes main (go-pkgs)

```
# main on master + linked on feature
worktree.List -> 2 entries (main + linked)
worktree.ListLinked -> 1 entry (linked only, IsMain=false)
```

## Preconditions

- Main repo with initial commit (helpers from root).

## Steps

1. Init main on `master`.
2. `git worktree add -b feature <linked> HEAD`.
3. Set `req.Op=list_and_linked` and fixture paths.

## Context

- Scenario 1 from P2: through go-pkgs `worktree.List` main+linked inventory.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	main := initRepo(t)
	linked := addLinkedBranch(t, main, "feature")
	req.Op = "list_and_linked"
	req.Dir = main
	req.MainPath = main
	req.LinkedPath = linked
	return nil
}
```
