# Scenario

**Feature**: FormatPlanPrompt lists CASE B commands with display formatting

```
MergeBack NeedsConfirm -> Confirm captures FormatPlanPrompt -> abort (no git mutations)
```

## Steps

- Set `CapturePrompt` and `DryRun=false`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CapturePrompt = true
	req.DryRun = false
	return nil
}
```