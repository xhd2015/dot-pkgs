# Scenario

**Feature**: git skip — a directory matched by a nested `.gitignore` (not the root) is pruned

```
# root/.gitignore has unrelated pattern; sub/.gitignore ignores build/ 
# sub/build/ contains untracked go.mod — must be skipped as gitignored
root + sub/.gitignore(ignored/) + sub/build/go.mod -> scan.Scan -> [.] (build/ absent)
```

This test exercises the nested `.gitignore` path in `shouldSkip`. Git reads `.gitignore`
files at every directory level, so `sub/.gitignore` is applied to paths under `sub/`.
The root's `git ls-files --others --ignored --exclude-standard --directory` (called by
`ListIgnoredDirs`) respects all nested `.gitignore` files — this test ensures a future
optimisation that removes the per-directory `git check-ignore` call does not regress
nested-gitignore handling.

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd.
2. Write root `.gitignore` with a different pattern (not matching `sub/`).
3. Create `sub/.gitignore` ignoring `build/`.
4. Create `sub/build/go.mod` (untracked and gitignored per nested `.gitignore`).
5. Set `req.RootDir` (operation `scan` is set by the `skips/` grouping Setup).

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initSkipRoot(t, "example.com/root")
	writeFile(t, filepath.Join(ws, ".gitignore"), "*.log\n")
	writeFile(t, filepath.Join(ws, "sub", ".gitignore"), "build/\n")
	writeModule(t, filepath.Join(ws, "sub", "build"), "example.com/root/sub/build")
	req.RootDir = ws
	return nil
}
```
