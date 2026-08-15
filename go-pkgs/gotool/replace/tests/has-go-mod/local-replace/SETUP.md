# Scenario

**Feature**: go.mod with local filesystem replace directives

```
# go.mod with local path replace -> NewVersion == "" -> file matched -> classify intra/extra repo
go.mod -> local filesystem replace -> classify -> intra or extra repo -> return issues
```

## Preconditions

- At least one go.mod has a local filesystem replace directive.

## Steps

1. Write go.mod files as specified by the leaf case.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// leaf cases write specific go.mod files
	return nil
}
```