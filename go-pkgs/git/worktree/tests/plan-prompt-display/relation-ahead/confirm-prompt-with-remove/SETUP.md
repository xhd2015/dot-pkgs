# Scenario

**Feature**: FormatPlanPrompt lists CASE B commands with display formatting

```
MergeBack NeedsConfirm -> Confirm captures FormatPlanPrompt -> abort (no git mutations)
```

## Steps

- Set `CapturePrompt` and `DryRun=false`.

```go
func Setup(t *testing.T, req *Request) error {
	req.CapturePrompt = true
	req.DryRun = false
	return nil
}
```