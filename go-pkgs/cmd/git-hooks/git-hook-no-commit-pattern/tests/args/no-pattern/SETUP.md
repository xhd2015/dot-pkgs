# Scenario

**Feature**: no patterns provided returns an error

```
# hook with no patterns -> error -> exit non-zero
hook binary -> no patterns -> "at least one pattern required"
```

## Preconditions

- The hook binary exists.
- No staged files are needed since the error occurs before file checking.

## Steps

1. Run the command with no positional arguments.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{}
	return nil
}
```
