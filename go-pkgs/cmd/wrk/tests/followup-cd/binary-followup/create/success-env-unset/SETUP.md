# Scenario

**Feature**: successful create without WRK_FOLLOWUP_FILE leaves no follow-up side effects

```
myrepo (main); WRK_FOLLOWUP_FILE unset
wrk -> stdout worktree path; pre-created followup.txt stays empty
```

## Steps

1. Init main repo.
2. Prepare empty follow-up path but do **not** export WRK_FOLLOWUP_FILE.
3. Run bare `wrk`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.UseFollowupEnv = false
	req.FollowupFile = defaultFollowupPath(req)
	req.CLIArgs = nil
	return nil
}
```
