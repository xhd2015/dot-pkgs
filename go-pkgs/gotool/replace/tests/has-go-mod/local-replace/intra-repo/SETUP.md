# Scenario

**Feature**: local replace target is inside the same git repo

```
# replace target path resolves to directory inside same git toplevel -> intra-repo
local replace -> resolve target -> git toplevel == top -> IsIntraRepo = true
```

## Preconditions

- The replace target directory exists and is inside the same git repo.

## Steps

1. Write go.mod with a local replace pointing to a directory inside the repo.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// leaf cases write specific local replace paths
	return nil
}
```