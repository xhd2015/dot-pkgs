# Scenario

**Feature**: WRK_AUTO_CD=0 disables wrapper follow-up and cd

```
WRK_AUTO_CD=0; source bash.sh; wrk from main
  -> worktree created; no stderr cd; FinalPWD stays main
```

## Steps

1. Init main repo.
2. Run wrapper create with `WRK_AUTO_CD=0`.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := setupMainRepo(t, req)
	req.RepoDir = mainRepo
	req.StartDir = mainRepo
	req.AutoCD = "0"
	req.CLIArgs = nil
	return nil
}
```
