# Scenario

**Feature**: no git at root — untracked-but-not-ignored dir with go.mod IS scanned

```
# root is NOT a git repo -> git-based skips disabled; only name skips (.git/vendor/testdata) apply
root(no .git) + untracked/go.mod -> scan.Scan -> [., untracked]  (untracked IS included)
```

When the root is not inside a git repo, the git-based skips (gitignore, nested-separate-repo)
are disabled. A plain untracked directory containing a `go.mod` is therefore scanned — only
the name-based skips (`.git`/`vendor`/`testdata`) can prune it.

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`).
2. Add `untracked/go.mod` (a nested module).
3. Do **NOT** init any git repo — there is no `.git` at root (or anywhere).
4. Set `req.RootDir` (operation `scan` is set by the `skips/` grouping Setup).

```go
func Setup(t *testing.T, req *Request) error {
	// Deliberately no git init (do NOT use initSkipRoot): root is not a git repo,
	// so git-based skips are disabled and untracked/ must be scanned.
	ws := newWorkspace(t)
	writeModule(t, ws, "example.com/root")
	writeModule(t, filepath.Join(ws, "untracked"), "example.com/root/untracked")
	req.RootDir = ws
	return nil
}
```
