# Scenario

**Feature**: local replace target is outside the git repo

```
# replace target path resolves to directory outside git toplevel or doesn't exist -> extra-repo
local replace -> resolve target -> git toplevel != top or not found -> IsIntraRepo = false
```

## Preconditions

- The replace target directory is outside the git repo or does not exist.

## Steps

1. Write go.mod with a local replace pointing outside the repo.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// leaf cases write specific local replace paths
	return nil
}
```