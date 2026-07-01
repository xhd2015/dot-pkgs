# Scenario

**Feature**: ScanStream emits modules in walk order (unsorted); Scan sorts

```
# stream leaves exercise the unsorted walk-order emission path
root + nested go.mod files -> scan.ScanStream -> per-module fn in walk order
```

## Steps

1. Leaf `Setup` creates an isolated workspace with the modules in the order the scenario
   requires (walk order depends on discovery sequence).
2. Init the workspace as a git repo (shared by all stream leaves).
3. Set `req.RootDir` (operation `stream` is set by this grouping Setup).

```go
// initStreamWorkspace creates an isolated workspace with a root go.mod at modulePath,
// inits it as a git repo, and returns the workspace dir. Shared by all stream leaves.
func initStreamWorkspace(t *testing.T, modulePath string) string {
	t.Helper()
	ws := newWorkspace(t)
	writeModule(t, ws, modulePath)
	initGitRepo(t, ws)
	return ws
}

func Setup(t *testing.T, req *Request) error {
	// stream leaves use the ScanStream (walk-order, unsorted) path
	req.Operation = "stream"
	return nil
}
```
