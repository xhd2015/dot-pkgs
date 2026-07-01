# Scenario

**Feature**: go.mod with version-only replaces (e.g. `old v1.0.0 => v2.0.0`)

```
# go.mod with version replace -> not local -> exit 0
go.mod -> scan -> module -> version-only replace (NewVersion != "") -> exit 0
```

## Preconditions

- A root go.mod exists with a version-based replace.

## Steps

1. Write `go.mod` with `replace example.com/old v1.0.0 => example.com/new v2.0.0`.

```go
import "os"
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	content := []byte("module example.com/myrepo\n\ngo 1.22\n\nrequire example.com/old v1.0.0\n\nreplace example.com/old v1.0.0 => example.com/new v2.0.0\n")
	return os.WriteFile(filepath.Join(req.RepoDir, "go.mod"), content, 0o644)
}

```
