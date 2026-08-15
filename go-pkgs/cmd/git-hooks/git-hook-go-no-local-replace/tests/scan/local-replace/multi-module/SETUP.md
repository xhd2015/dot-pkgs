# Scenario

**Feature**: multiple go.mod files with an intra-repo local replace

```
# multiple go.mod files -> scan -> intra-repo local replaces are allowed by default -> exit 0
root go.mod (no local) + sub/go.mod (local ./local) -> scan -> intra-repo -> exit 0
```

## Preconditions

- Root go.mod has no local replace.
- A subdirectory go.mod has a local replace to an existing directory inside the same git repo.

## Steps

1. Write root `go.mod` with no replaces.
2. Write `sub/go.mod` with a local replace.
3. Create the local target directory.

```go
import (
	"os"

	"github.com/xhd2015/doctest/session"
)
import "path/filepath"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// root go.mod — no local replace
	if err := os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), []byte("module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old v0.0.0 => example.com/new v1.0.0\n"), 0o644); err != nil {
		return err
	}
	// sub/go.mod — local replace
	subDir := filepath.Join(req.RepoDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(subDir, "go.mod"), []byte("module example.com/sub\n\ngo 1.22\n\nrequire example.com/other v0.0.0\n\nreplace example.com/other => ./local\n"), 0o644); err != nil {
		return err
	}
	localDir := filepath.Join(subDir, "local")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(localDir, "go.mod"), []byte("module example.com/local\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}
	return nil
}

```
