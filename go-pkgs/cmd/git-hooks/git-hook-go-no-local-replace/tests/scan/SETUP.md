# Scenario

**Feature**: scan finds local path replaces in go.mod files

```
# go.mod files in repo -> scan -> replace inspection -> local replace detection -> output
go.mod files -> module list -> replace check -> local? -> print + exit 1 | exit 0
```

## Preconditions

- No domain filter is set (hook runs unconditionally).

## Steps

1. Write go.mod files as specified by the leaf case.
2. Run the hook binary with no domain filter.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = nil // no domain filter, scan all modules
	return nil
}

```
