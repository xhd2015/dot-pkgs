# Scenario

**Feature**: create with --no-cd suppresses follow-up write even when env set

```
myrepo + WRK_FOLLOWUP_FILE=tmp
wrk --no-cd -> worktree created; follow-up file empty
```

## Steps

1. Init main repo.
2. Run `wrk --no-cd` with follow-up env set.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--no-cd"}
	return nil
}
```
