# Scenario

**Feature**: basic Scan over a multi-module workspace returns every module sorted by Dir

```
# root go.mod + nested go.mod files under one git repo
root + nested go.mod files -> scan.Scan -> sorted []Module
```

## Steps

1. Create an isolated workspace with a root `go.mod`.
2. Init the workspace as a git repo (shared by all basic leaves).
3. Leaf `Setup` adds the nested modules specific to the scenario and sets `req.Operation`.

```go
// initBasicWorkspace creates an isolated workspace with a root go.mod at modulePath,
// inits it as a git repo, and returns the workspace dir. Shared by all basic leaves.
import "github.com/xhd2015/doctest/session"

func initBasicWorkspace(t *testing.T, modulePath string) string {
	t.Helper()
	ws := newWorkspace(t)
	writeModule(t, ws, modulePath)
	initGitRepo(t, ws)
	return ws
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// basic leaves use the sorted-batch Scan path
	req.Operation = "scan"
	return nil
}
```
