# Scenario

**Feature**: attached branch ahead dry-run uses branch name in merge

```
# worktree on branch feature (not detached)
MergeBack DryRun=true -> merge --ff-only feature (not commit hash)
```

## Steps

- Set `DryRun=true`, `CapturePrompt=false`.

```go
func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.CapturePrompt = false
	return nil
}
```