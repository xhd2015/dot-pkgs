# Scenario

**Feature**: `--strict` flag blocks local replaces including intra-repo

```
# --strict flag -> intra-repo replace -> blocked -> exit 1
hook binary --strict -> go.mod with ./sub (intra-repo) -> exit 1
```

## Preconditions

- A root go.mod exists with `replace example.com/old => ./sub`.
- The `./sub` directory exists inside the repo (intra-repo).
- The `--strict` flag is passed.

## Steps

1. Write `go.mod` with a `./` local replace.
2. Create the `./sub/go.mod` target inside the repo.
3. Run the hook with `--strict`.

```go
import "os"
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--strict"}
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