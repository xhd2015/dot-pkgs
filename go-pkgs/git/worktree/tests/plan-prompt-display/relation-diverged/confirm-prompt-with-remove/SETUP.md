# Scenario

**Feature**: diverged confirm prompt shows rebase comment and full command set

```
Confirm captures FormatPlanPrompt for relation diverged
```

```go
func Setup(t *testing.T, req *Request) error {
	req.CapturePrompt = true
	req.DryRun = false
	return nil
}
```