# Scenario

**Feature**: detached HEAD ahead dry-run lists merge with commit hash

```
MergeBack DryRun=true -> printDryRun -> merge --ff-only <sha> (not HEAD or branch name)
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