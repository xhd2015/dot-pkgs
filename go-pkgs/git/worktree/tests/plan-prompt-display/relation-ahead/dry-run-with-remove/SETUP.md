# Scenario

**Feature**: printDryRun uses the same command display formatter

```
MergeBack DryRun=true -> printDryRun -> stdout command lines (no prompt header)
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