# Scenario

**Feature**: main checkout status block omits Compare with Master

```
# no linked worktrees registered
myrepo (main only) -> wrk --status -> root block without Compare with Master
```

## Steps

1. Initialize `{WorkRoot}/myrepo` on branch `main`.
2. Run `wrk --status` from the main repo root.

```go
func Setup(t *testing.T, req *Request) error {
	ensureCompareWithMasterHelpersUsed()
	mainRepo := setupMainRepoWithSubject(t, req.WorkRoot, "myrepo", "status main root")

	req.RepoDir = mainRepo
	req.MainRepo = mainRepo
	return nil
}
```