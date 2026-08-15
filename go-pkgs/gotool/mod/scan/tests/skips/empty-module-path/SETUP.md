# Scenario

**Feature**: go.mod with no module path is a boundary only — not emitted as a Module

```
# comment-only / empty go.mod under boundary/ — parent module boundary, no Path
root + boundary/go.mod(no module line) -> scan.Scan -> [.]  (boundary absent)
```

Boundary stubs (e.g. codegen template dirs) keep packages out of the parent module
without being publishable modules themselves.

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd.
2. Add `boundary/go.mod` with comments only (no `module` line).
3. Set `req.RootDir`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initSkipRoot(t, "example.com/root")
	boundary := filepath.Join(ws, "boundary")
	if err := os.MkdirAll(boundary, 0o755); err != nil {
		t.Fatalf("mkdir boundary: %v", err)
	}
	writeFile(t, filepath.Join(boundary, "go.mod"), "// Not a Go module.\n// Boundary only.\n")
	req.RootDir = ws
	return nil
}
```
