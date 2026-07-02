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
func Setup(t *testing.T, req *Request) error {
	// leaf cases write specific local replace paths
	_ = t
	return nil
}
```