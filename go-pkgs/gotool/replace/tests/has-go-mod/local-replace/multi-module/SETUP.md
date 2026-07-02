# Scenario

**Feature**: multiple go.mod files across the repo

```
# multiple go.mod files -> scan finds all -> each module checked independently -> accumulate issues
multiple go.mod files -> scan each module -> check replaces independently -> all issues returned
```

## Preconditions

- Multiple go.mod files exist in the repo.
- At least one has a local replace.

## Steps

1. Write multiple go.mod files as specified by the leaf case.

```go
func Setup(t *testing.T, req *Request) error {
	// leaf cases write specific go.mod files
	_ = t
	return nil
}
```