# Scenario

**Feature**: wrk -v create streams worktree add output on branch-collision --no-checkout path

```
# branch main-{date} pre-exists; fixed <target-dir> reuses branch via worktree add --no-checkout
myrepo (main) + refs/heads/main-2026-06-30 -> wrk myrepo <wt> -v -> stderr streams add output only
```

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "myrepo")
	initFetchVerboseRepo(t, repo, "create v branch collision no-checkout")
	runGitIsolated(t, repo, "branch", branchName("main", wrkDate, 0))
	req.TargetDir = repo
	req.RepoDir = req.WorkRoot
	req.SpawnDir = filepath.Join(req.WorkRoot, "wt")
	req.Args = []string{"-v"}
	return nil
}
```