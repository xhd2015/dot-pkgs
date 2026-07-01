# Scenario

**Feature**: go.mod with `./` local replace

```
# replace old => ./local -> local path (NewVersion == "") -> print -> exit 1
go.mod -> scan -> replace old => ./local -> local -> print "./local" -> exit 1
```

## Preconditions

- A root go.mod exists with `replace example.com/old => ./local`.

## Steps

1. Write `go.mod` with a `./` local replace.
2. Create the `./local/go.mod` so the scan doesn't error on the directory.

```go
import "os"
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	if err := os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), []byte("module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => ./local\n"), 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(req.RepoDir, "local"), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(req.RepoDir, "local", "go.mod"), []byte("module example.com/local\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}
	return nil
}

```
