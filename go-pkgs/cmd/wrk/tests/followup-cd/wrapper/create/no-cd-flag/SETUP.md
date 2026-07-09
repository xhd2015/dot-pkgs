# Scenario

**Feature**: wrapper respects --no-cd (no auto-cd)

```
source bash.sh; wrk --no-cd from main
  -> worktree created; no stderr cd; FinalPWD stays main
```

## Steps

1. Init main repo.
2. Run `wrk --no-cd` via wrapper.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = mainRepo
	req.CLIArgs = []string{"--no-cd"}
	return nil
}
```
