# Scenario

**Feature**: intra-repo local replace is allowed by default (lenient behavior)

```
# replace old => ./sub -> ./sub inside same repo -> intra-repo -> allowed -> exit 0
go.mod -> scan -> replace old => ./sub -> intra-repo -> exit 0 (NEW behavior)
```

## Preconditions

- A root go.mod exists with `replace example.com/old => ./sub`.
- The `./sub` directory exists and contains a go.mod inside the same git repo.
- No `--strict` flag is set.

## Steps

1. Write `go.mod` with a `./` local replace to an intra-repo target.
2. Create the `./sub/go.mod` target directory inside the repo.
3. Run the hook with default (lenient) mode.

```go
import (
	"os"

	"github.com/xhd2015/doctest/session"
)
import "path/filepath"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = nil // default lenient mode, no --strict
	if err := os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), []byte("module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ./sub\n"), 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(req.RepoDir, "sub"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(req.RepoDir, "sub", "go.mod"), []byte("module example.com/sub\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}
	return nil
}
```