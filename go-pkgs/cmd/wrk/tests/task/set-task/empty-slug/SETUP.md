# Scenario

**Feature**: --set-task where description slugifies to empty errors

```
wrk --set-task "!!!" -> slug empty -> non-zero exit, error
```

## Steps

1. Create a worktree.
2. Set req.SetTaskDesc = "!!!".
3. Verify non-zero exit — slug is empty.

```go
func Setup(t *testing.T, req *Request) error {
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGit(t, mainRepo, "add", "go.mod")
	runGit(t, mainRepo, "commit", "-m", "add go.mod")

	wtDir := runWrkWithArgs(t, req, mainRepo, "--task", "original task")
	req.WtDir = wtDir
	req.RepoDir = wtDir
	req.SetTaskDesc = "!!!"
	return nil
}
```