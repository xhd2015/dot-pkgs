# Scenario

**Feature**: skip rules prune subtrees during Scan; each leaf exercises one skip class

```
# scan applies name skips (.git/vendor/testdata) and, when root is a git repo, git skips
root + (skip-target dir with go.mod) -> scan.Scan -> [.]  (skip-target absent)
```

The `skips/` siblings are MECE over the skip rule exercised: name-based (`testdata`,
`vendor`), gitignore, nested-separate-repo, and the negative case (no git at root →
untracked dir IS scanned). All leaves use the sorted-batch `Scan` path.

## Steps

1. Leaf `Setup` creates an isolated workspace with a root `go.mod`.
2. Leaf adds the skip-target directory (and git setup / `.gitignore` as needed).
3. Set `req.RootDir` (operation `scan` is set by this grouping Setup).

```go
// initSkipRoot creates an isolated workspace with a root go.mod (modulePath) and inits it
// as a git repo, returning the workspace dir. Used by skip leaves whose root is a git repo.
import "github.com/xhd2015/doctest/session"

func initSkipRoot(t *testing.T, modulePath string) string {
	t.Helper()
	ws := newWorkspace(t)
	writeModule(t, ws, modulePath)
	initGitRepo(t, ws)
	return ws
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// all skip leaves use the sorted-batch Scan path
	req.Operation = "scan"
	return nil
}
```
