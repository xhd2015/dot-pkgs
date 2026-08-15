# Scenario

**Feature**: multiple modules, only one has a local replace

```
# root go.mod (no local) + sub/go.mod (local ./local) -> only sub has issue
root go.mod (clean) + sub/go.mod (replace ./local) -> 1 issue in sub
```

## Preconditions

- Root go.mod has no local replace.
- A subdirectory go.mod has a local replace.

## Steps

1. Write root `go.mod` with a version-only replace (no local).
2. Write `sub/go.mod` with a local replace.
3. Create the local target directory.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// root go.mod — no local replace
	if err := writeGoMod(req.RootDir, "go.mod", "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old v0.0.0 => example.com/new v1.0.0\n"); err != nil {
		return err
	}
	// sub/go.mod — local replace
	if err := writeGoMod(req.RootDir, "sub/go.mod", "module example.com/sub\n\ngo 1.22\n\nrequire example.com/other v0.0.0\n\nreplace example.com/other => ./local\n"); err != nil {
		return err
	}
	// Create the local target inside sub/ (intra-repo)
	if err := writeGoMod(req.RootDir, "sub/local/go.mod", "module example.com/local\n\ngo 1.22\n"); err != nil {
		return err
	}
	return nil
}
```