# Scenario

**Feature**: argument parsing for git-hook-go-no-local-replace

```
# hook binary with args -> parse -> output / error
hook binary --flag -> parseArgs -> result
```

## Preconditions

- The hook binary exists.

## Steps

1. Run the hook binary with specific argument flags.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = nil // leaf cases set specific args
	return nil
}

```
