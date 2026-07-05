# Scenario

**Feature**: --color highlights Needs Push remote summary in orange

```
main ahead of origin/main -> wrk --projects --color -> orange Needs Push(...)
```

## Steps

1. Create tracked repo `{WorkRoot}/push` pushed to bare `origin`.
2. Commit once more on `main` (1 commit ahead).
3. Record and run `wrk --projects --color`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	ensureColorOutputHelpersUsed()
	withProjectsColor(req)
	origin := setupColorBareOrigin(t, req.WorkRoot, "origin")
	repo := setupColorTrackedMainRepo(t, req.WorkRoot, "push", origin, "push base")
	writeFile(t, filepath.Join(repo, "ahead.txt"), "ahead\n")
	runGit(t, repo, "add", "ahead.txt")
	runGit(t, repo, "commit", "-m", "ahead of upstream")
	recordColorProject(t, req, repo)
	req.MainRepo = repo
	return nil
}
```