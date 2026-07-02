# Scenario

**Feature**: `--strict` flag blocks intra-repo replaces in multi-module setup

```
# --strict flag -> multiple modules, sub has intra-repo replace -> blocked -> exit 1
hook binary --strict -> root clean + sub/go.mod (./sub intra-repo) -> exit 1
```

## Preconditions

- Root go.mod has no local replace.
- A subdirectory go.mod has an intra-repo local replace.
- The `--strict` flag is passed.

## Steps

1. Write root `go.mod` with version-only replace (no local).
2. Write `sub/go.mod` with an intra-repo local replace.
3. Create the target directory inside the repo.
4. Run the hook with `--strict`.

```go
import "os"
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--strict"}
	// root go.mod — no local replace
	if err := os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), []byte("module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old v0.0.0 => example.com/new v1.0.0\n"), 0o644); err != nil {
		return err
	}
	// sub/go.mod — intra-repo local replace
	subDir := filepath.Join(req.RepoDir, "sub")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(subDir, "go.mod"), []byte("module example.com/sub\n\ngo 1.22\n\nrequire example.com/other v0.0.0\n\nreplace example.com/other => ./local\n"), 0o644); err != nil {
		return err
	}
	// Create the local target inside sub/ (intra-repo)
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