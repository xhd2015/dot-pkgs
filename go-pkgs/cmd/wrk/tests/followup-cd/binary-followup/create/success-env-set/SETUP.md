# Scenario

**Feature**: successful create with WRK_FOLLOWUP_FILE writes cd to worktree

```
myrepo (main) + WRK_FOLLOWUP_FILE=tmp
wrk -> stdout worktree path; follow-up: cd <abs-worktree>
```

## Steps

1. Init main repo `myrepo`.
2. Run bare `wrk` with follow-up env set.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.UseFollowupEnv = true
	req.CLIArgs = nil // bare create
	return nil
}
```
