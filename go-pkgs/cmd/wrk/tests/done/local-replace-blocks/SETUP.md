# Scenario

**Feature**: wrk --done errors when consumer go.mod has filesystem replace

```
# linked wt with replace => ./external/foo -> wrk --done -> guard error before parent merge-back
```

## Steps

1. Create main repo and consumer linked worktree via `wrk`.
2. Add `replace example.com/foo => ./external/foo` to consumer go.mod.
3. Run `wrk --done` from worktree.

```go
import (
	"os/exec"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	_, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	writeFile(t, filepath.Join(wtDir, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	mkdirAll(t, filepath.Join(wtDir, "external", "foo"))
	runGoMod(t, wtDir, "edit", "-replace=example.com/foo=./external/foo")

	req.RepoDir = wtDir
	req.WtBranch = branch
	req.Args = []string{"--done"}
	return nil
}

func runGoMod(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}
```