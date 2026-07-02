# Scenario

**Feature**: absolute-path replace target inside same repo

```
# replace old => /abs/path/inside/repo -> target inside top -> same git toplevel -> intra-repo
go.mod -> replace old => <abs path in repo> -> target inside repo -> IsIntraRepo = true
```

## Preconditions

- A root go.mod exists with an absolute-path replace pointing to a directory inside the same git repo.

## Steps

1. Write `go.mod` with an absolute-path local replace to a directory inside the repo.
2. Create the target directory inside the repo.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	targetDir := filepath.Join(req.RootDir, "internal", "pkg")
	if err := writeGoMod(req.RootDir, "internal/pkg/go.mod", "module example.com/internal-pkg\n\ngo 1.22\n"); err != nil {
		return err
	}
	content := "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => " + targetDir + "\n"
	return writeGoMod(req.RootDir, "go.mod", content)
}
```