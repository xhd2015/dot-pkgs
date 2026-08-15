# Scenario

**Feature**: git skip — a directory matched by the root `.gitignore` is pruned

```
# .gitignore lists ignored/; scan must treat gitignored dirs as non-part-of-workspace
root + .gitignore(ignored/) + ignored/go.mod -> scan.Scan -> [.]  (ignored absent)
```

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd.
2. Write `.gitignore` containing `ignored/`.
3. Add `ignored/go.mod` (ignored by git per `.gitignore`, so it stays untracked).
4. Set `req.RootDir` (operation `scan` is set by the `skips/` grouping Setup).

Note: git applies `.gitignore` to working-tree paths whether or not the `.gitignore` itself
is committed, so the ignore rule is effective during scan.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initSkipRoot(t, "example.com/root")
	writeFile(t, filepath.Join(ws, ".gitignore"), "ignored/\n")
	writeModule(t, filepath.Join(ws, "ignored"), "example.com/root/ignored")
	req.RootDir = ws
	return nil
}
```
