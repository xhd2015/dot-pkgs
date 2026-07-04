# Scenario

**Feature**: nested independent main repos omit Compare with Master

```
# root checkout + nested independent repo (multiple-git-dirs pattern)
myrepo + myrepo/tools/child -> wrk --status -> neither block has Compare with Master
```

## Steps

1. Initialize `{WorkRoot}/myrepo` as a git repo on branch `main`.
2. Initialize `{WorkRoot}/myrepo/tools/child` as an independent git repo on branch `main`.
3. Run `wrk --status` from `{WorkRoot}/myrepo`.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	ensureCompareWithMasterHelpersUsed()
	repo := filepath.Join(req.WorkRoot, "myrepo")
	child := filepath.Join(repo, "tools", "child")

	statusInitRepoWithSubject(t, repo, "root status repo")
	statusInitRepoWithSubject(t, child, "child status repo")

	req.RepoDir = repo
	req.MainRepo = repo
	req.DepPath = child
	return nil
}
```