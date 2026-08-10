# Scenario

**Feature**: repo has at least one go.mod file

```
# go.mod exists -> scan finds module(s) -> inspect replaces
top -> scan -> modules found -> check replace directives -> issues or nil
```

## Preconditions

- The repo root contains at least one go.mod file.

## Steps

1. Write go.mod files as specified by the leaf case.

```go
func Setup(t *testing.T, req *Request) error {
	// leaf cases write specific go.mod files
	return nil
}
```