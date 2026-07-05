# Scenario

**Feature**: branch ref collision skips date-suffixed attempt

```
# branch main-{date} pre-exists while cwd is on main; then wrk
myrepo (main) + refs/heads/main-2026-06-30 -> wrk -> myrepo-main-2026-06-30-1
```

## Steps

1. Initialize git repo `myrepo` on branch `main`.
2. Pre-create branch `main-2026-06-30` with `git branch` (ref exists, date-suffixed worktree path still free).
3. Run `wrk` from `myrepo` on `main`.

```go
func Setup(t *testing.T, req *Request) error {
	initGitRepoOnMain(t, req.RepoDir)
	runGitIsolated(t, req.RepoDir, "branch", branchName("main", wrkDate, 0))
	return nil
}
```