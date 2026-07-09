# Scenario

**Feature**: wrapper create auto-cds into new worktree

```
source bash.sh from main repo
wrk -> stderr "cd <worktree>"; FinalPWD = worktree; exit 0
```

## Steps

1. Init main repo; start shell cwd there.
2. Invoke `wrk` via installed wrapper (default auto-cd on).

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = mainRepo
	req.CLIArgs = nil
	return nil
}
```
