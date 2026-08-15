# Scenario

**Feature**: no go.mod files exist in the repository

```
# empty repo -> no go.mod files -> scan finds nothing -> exit 0
empty repo -> scan -> no modules -> exit 0
```

## Preconditions

- No go.mod files are written to the repository.

## Steps

1. Run the hook binary.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = nil // no go.mod files, hook runs with no args
	return nil
}

```
