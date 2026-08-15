# Scenario

**Feature**: go.mod with exclusively version-based replace directives

```
# replace old v0.0.0 => new v1.0.0 -> version replace (NewVersion != "") -> not local -> nil issues
go.mod -> version replaces -> not local filesystem -> nil issues
```

## Preconditions

- A root go.mod exists with only version-based replace directives.

## Steps

1. Write `go.mod` with a version-based replace directive.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	content := "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old v0.0.0 => example.com/new v1.0.0\n"
	return writeGoMod(req.RootDir, "go.mod", content)
}
```