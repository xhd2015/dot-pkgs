# Scenario

**Feature**: `../` replace target inside same repo

```
# replace old => ../sub -> ../sub exists inside top -> same git toplevel -> intra-repo
go.mod -> replace old => ../sub -> target still in same repo -> IsIntraRepo = true
```

## Preconditions

- A go.mod exists in a subdirectory with `replace example.com/old => ../sibling`.
- The `../sibling` directory exists inside the same git repo.

## Steps

1. Create `sub/go.mod` with a `../` local replace.
2. Create `sibling/go.mod` as the target directory inside the same repo.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	// Root go.mod
	if err := writeGoMod(req.RootDir, "go.mod", "module example.com/myrepo\n\ngo 1.22\n"); err != nil {
		return err
	}
	// sub/go.mod with ../sibling replace
	subContent := "module example.com/sub\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ../sibling\n"
	if err := writeGoMod(req.RootDir, "sub/go.mod", subContent); err != nil {
		return err
	}
	// sibling/go.mod inside the same repo
	if err := writeGoMod(req.RootDir, "sibling/go.mod", "module example.com/sibling\n\ngo 1.22\n"); err != nil {
		return err
	}
	return nil
}
```