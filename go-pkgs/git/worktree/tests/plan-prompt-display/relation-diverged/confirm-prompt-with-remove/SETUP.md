# Scenario

**Feature**: diverged confirm prompt shows rebase comment and full command set

```
Confirm captures FormatPlanPrompt for relation diverged
```

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.CapturePrompt = true
	req.DryRun = false
	return nil
}
```