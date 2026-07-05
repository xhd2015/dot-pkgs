# Scenario

**Feature**: --color highlights Needs Pull remote summary in orange

```
main behind origin/main -> wrk --projects --color -> orange Needs Pull
```

## Steps

1. Create tracked repo `{WorkRoot}/pull` pushed to bare `origin`.
2. Push an additional commit to `origin/main` from a clone (main stays behind).
3. Record and run `wrk --projects --color`.

```go
func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "pull", origin, "pull base")
	pushCommitToBareOrigin(t, req.WorkRoot, origin, "remote-only.txt", "remote\n", "on origin only")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```