# Scenario

**Feature**: name skip — a `testdata/` directory is pruned, its go.mod never scanned

```
# testdata is a reserved Go fixture dir; scan must prune it like the existing kool walk
root + testdata/x/go.mod -> scan.Scan -> [.]  (testdata subtree absent)
```

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd.
2. Add `testdata/x/go.mod` (which would otherwise be a nested module).
3. Set `req.RootDir` (operation `scan` is set by the `skips/` grouping Setup).

```go
func Setup(t *testing.T, req *Request) error {
	ws := initSkipRoot(t, "example.com/root")
	writeModule(t, filepath.Join(ws, "testdata", "x"), "example.com/root/testdata-x")
	req.RootDir = ws
	return nil
}
```
