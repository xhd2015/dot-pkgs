# Scenario

**Feature**: git skip — a nested separate git repo (own `.git`, not a submodule) is pruned

```
# ext/ has its OWN .git from a separate `git init ext`; root never tracked it as a submodule
root + ext/go.mod + ext/.git(separate repo, untracked by root) -> scan.Scan -> [.]  (ext absent)
```

The `ext/` directory is a real git repo in its own right but is NOT a submodule of the
root (no gitlink in the root index). The scan must treat it as a nested separate repo and
skip the whole subtree — `ext`'s own `go.mod` is never scanned.

## Steps

1. Create an isolated workspace with root `go.mod` (`example.com/root`), git-init'd
   (root committed before `ext/` exists, so `ext/` is never in the root index).
2. Create `ext/go.mod` (`example.com/ext`) and init `ext/` as its **own** git repo
   (separate `.git` inside `ext/`, NOT a submodule of root).
3. Set `req.RootDir` (operation `scan` is set by the `skips/` grouping Setup).

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	ws := initSkipRoot(t, "example.com/root")

	// ext/ is its own separate git repo, created after root's commit so root never
	// tracks it (no gitlink in the root index -> not a submodule).
	ext := filepath.Join(ws, "ext")
	writeModule(t, ext, "example.com/ext")
	initGitRepo(t, ext)

	req.RootDir = ws
	return nil
}
```
