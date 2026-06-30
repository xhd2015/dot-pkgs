# Scenario

**Feature**: slash in branch name sanitized for path token only

```
# branch feature/foo keeps slash in git branch; path uses feature-foo token
myrepo (feature/foo) -> wrk -> myrepo-feature-foo-2026-06-30
```

## Steps

1. Initialize git repo `myrepo` on branch `main`.
2. Create and check out branch `feature/foo`.
3. Run `wrk` from `myrepo`.

```go
func Setup(t *testing.T, req *Request) error {
	initGitRepoOnMain(t, req.RepoDir)
	runGit(t, req.RepoDir, "checkout", "-b", "feature/foo")
	return nil
}
```