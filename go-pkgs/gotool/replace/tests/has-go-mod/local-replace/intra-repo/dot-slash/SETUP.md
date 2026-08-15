# Scenario

**Feature**: `./` replace target inside same repo

```
# replace old => ./sub -> ./sub exists inside top -> same git toplevel -> intra-repo
go.mod -> replace old => ./sub -> target inside repo -> IsIntraRepo = true
```

## Preconditions

- A root go.mod exists with `replace example.com/old => ./sub`.
- The `./sub` directory exists and contains a go.mod inside the same git repo.

## Steps

1. Write `go.mod` with a `./` local replace.
2. Create the `./sub/go.mod` target directory inside the repo.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	content := "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ./sub\n"
	if err := writeGoMod(req.RootDir, "go.mod", content); err != nil {
		return err
	}
	if err := writeGoMod(req.RootDir, "sub/go.mod", "module example.com/sub\n\ngo 1.22\n"); err != nil {
		return err
	}
	return nil
}
```