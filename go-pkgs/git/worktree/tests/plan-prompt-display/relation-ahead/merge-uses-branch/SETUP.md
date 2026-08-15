# Scenario

**Feature**: attached branch ahead dry-run uses branch name in merge

```
# worktree on branch feature (not detached)
MergeBack DryRun=true -> merge --ff-only feature (not commit hash)
```

## Steps

- Set `DryRun=true`, `CapturePrompt=false`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DryRun = true
	req.CapturePrompt = false
	return nil
}
```