# Scenario

**Feature**: go.mod with absolute-path local replace

```
# replace old => /home/user/pkg -> local path (NewVersion == "") -> print -> exit 1
go.mod -> scan -> replace old => /tmp/somepkg -> local -> print "/tmp/somepkg" -> exit 1
```

## Preconditions

- A root go.mod exists with `replace example.com/old => /tmp/somepkg`.

## Steps

1. Write `go.mod` with an absolute-path local replace.
2. Create the target directory.

```go
import "os"
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	if err := os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), []byte("module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v0.0.0\n\nreplace example.com/old => /tmp/somepkg\n"), 0o644); err != nil {
		return err
	}
	if err := os.MkdirAll("/tmp/somepkg", 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join("/tmp/somepkg", "go.mod"), []byte("module example.com/somepkg\n\ngo 1.22\n"), 0o644); err != nil {
		return err
	}
	return nil
}

```
