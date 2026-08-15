# Scenario

**Feature**: argument parsing for git-hook-no-commit-pattern

```
# hook binary with args -> parse -> output / error
hook binary --flag -> parseArgs -> result
```

## Preconditions

- The hook binary exists.

## Steps

1. Run the hook binary with specific argument flags.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = nil // leaf cases set specific args
	return nil
}
```
