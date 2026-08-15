# Scenario

**Feature**: go.mod with `../` local replace

```
# replace old => ../sibling -> local path (NewVersion == "") -> print -> exit 1
go.mod -> scan -> replace old => ../another -> local -> print "../another" -> exit 1
```

## Preconditions

- A root go.mod exists with `replace example.com/old => ../another`.

## Steps

1. Write `go.mod` with a `../` local replace.
2. Create the sibling directory with a go.mod.

```go
import (
	"os"

	"github.com/xhd2015/doctest/session"
)
import "path/filepath"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), []byte("module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ../another\n"), 0o644); err != nil {
		return err
	}
	// The sibling dir is outside the repo, create it at the git worktree root level
	anotherDir := filepath.Join(filepath.Dir(req.RepoDir), "another")
	if err := os.MkdirAll(anotherDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(anotherDir, "go.mod"), []byte("module example.com/another\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}
	return nil
}

```
