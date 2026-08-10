# Scenario

**Feature**: go.mod with only version-based replace directives

```
# go.mod has replaces but all version-based -> NewVersion != "" -> not local -> nil issues
go.mod -> version replaces only -> not local filesystem -> nil issues
```

## Preconditions

- A root go.mod exists with only version-based replace directives.

## Steps

1. Write go.mod files as specified by the leaf case.

```go
func Setup(t *testing.T, req *Request) error {
	// leaf cases write specific go.mod files
	return nil
}
```