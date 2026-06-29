# Scenario

**Feature**: printDryRun uses the same command display formatter

```
MergeBack DryRun=true -> printDryRun -> stdout command lines (no prompt header)
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