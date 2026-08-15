# Scenario

**Feature**: no go.mod files in the repo

```
# no go.mod files -> scan finds nothing -> nil issues
top -> scan -> no modules -> nil issues
```

## Preconditions

- The repo root contains no go.mod files.

## Steps

1. Leave the repo empty (no go.mod files).

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// repo is empty (git init already done by root SETUP), no go.mod files
	return nil
}
```