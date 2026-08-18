# Scenario

**Feature**: local replace target is outside the git repo

```
# replace target path resolves outside the scanning worktree tree -> extra-repo
local replace -> resolve target -> path not under scan root -> IsIntraRepo = false
```

## Preconditions

- The replace target path is outside the scanning worktree tree.

## Steps

1. Write go.mod with a local replace pointing outside the repo.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// leaf cases write specific local replace paths
	return nil
}
```