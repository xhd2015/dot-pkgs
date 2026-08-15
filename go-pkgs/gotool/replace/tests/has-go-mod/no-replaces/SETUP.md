# Scenario

**Feature**: go.mod with no replace directives

```
# go.mod exists but no replace directives -> no local replaces -> nil issues
go.mod (no replace) -> scan -> module has no local replaces -> nil issues
```

## Preconditions

- A root go.mod exists with no replace directives.

## Steps

1. Write `go.mod` with a module and require but no replace directives.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	content := "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/other v1.0.0\n"
	return writeGoMod(req.RootDir, "go.mod", content)
}
```