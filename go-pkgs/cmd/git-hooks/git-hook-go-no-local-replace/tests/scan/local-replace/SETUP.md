# Scenario

**Feature**: go.mod with local-path replace directives

```
# go.mod with local path replace -> NewVersion == "" -> local -> print + exit 1
go.mod -> scan -> module -> local replace (NewVersion == "") -> print path -> exit 1
```

## Preconditions

- The leaf case writes a go.mod with a specific local replace type.

## Steps

1. Write go.mod with a local-path replace as specified by the leaf.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = nil // leaf cases write specific go.mod files
	return nil
}

```
