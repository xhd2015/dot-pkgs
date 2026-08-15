# Scenario

**Feature**: go.mod with no replace directives

```
# go.mod with module path only -> no replace -> exit 0
go.mod -> scan -> module -> no replaces -> exit 0
```

## Preconditions

- A root go.mod exists with no replace directives.

## Steps

1. Write `go.mod` with only a module declaration.

```go
import (
	"os"

	"github.com/xhd2015/doctest/session"
)
import "path/filepath"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	content := []byte("module example.com/myrepo\n\ngo 1.22\n")
	return os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), content, 0o644)
}

```
