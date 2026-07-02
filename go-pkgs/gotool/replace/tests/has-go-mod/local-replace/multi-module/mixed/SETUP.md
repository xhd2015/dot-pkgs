# Scenario

**Feature**: multiple modules with mixed intra-repo and extra-repo replaces

```
# root go.mod (intra-repo ./sub) + sub2/go.mod (extra-repo ./nonexistent) -> both found
root go.mod (intra-repo) + sub2/go.mod (extra-repo) -> 2 issues with mixed IsIntraRepo flags
```

## Preconditions

- Root go.mod has an intra-repo local replace.
- A subdirectory go.mod has an extra-repo local replace.

## Steps

1. Write root `go.mod` with a local replace to `./sub` (intra-repo, target exists in repo).
2. Write `sub2/go.mod` with a local replace to `./nonexistent` (extra-repo, target doesn't exist).
3. Create the `./sub/go.mod` target for the intra-repo case.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	// root go.mod — intra-repo replace to ./sub
	if err := writeGoMod(req.RootDir, "go.mod", "module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ./sub\n"); err != nil {
		return err
	}
	// Create sub/go.mod so the target exists and is intra-repo
	if err := writeGoMod(req.RootDir, "sub/go.mod", "module example.com/sub\n\ngo 1.22\n"); err != nil {
		return err
	}
	// sub2/go.mod — extra-repo replace to ./nonexistent
	if err := writeGoMod(req.RootDir, "sub2/go.mod", "module example.com/sub2\n\ngo 1.22\n\nrequire example.com/other v0.0.0\n\nreplace example.com/other => ./nonexistent\n"); err != nil {
		return err
	}
	return nil
}
```