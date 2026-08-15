# Scenario

**Feature**: absolute-path replace target outside git repo

```
# replace old => /tmp/outside -> /tmp/outside outside repo -> different git toplevel -> extra-repo
go.mod -> replace old => /tmp/outside -> target outside repo -> IsIntraRepo = false
```

## Preconditions

- A root go.mod exists with an absolute-path replace to a directory outside the git repo.

## Steps

1. Write `go.mod` with an absolute-path local replace to an external directory.
2. Create the target directory at an external location.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	extDir := filepath.Join(t.TempDir(), "outside-pkg")
	if err := writeGoMod(extDir, "go.mod", "module example.com/outside\n\ngo 1.22\n"); err != nil {
		return err
	}
	// Initialize separate git repo in the external directory
	if err := runGit(extDir, "init"); err != nil {
		return err
	}
	content := "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => " + extDir + "\n"
	return writeGoMod(req.RootDir, "go.mod", content)
}
```