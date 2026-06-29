# Scenario

**Feature**: diverged dry-run command listing

```go
func Setup(t *testing.T, req *Request) error {
	req.DryRun = true
	req.CapturePrompt = false
	return nil
}
```