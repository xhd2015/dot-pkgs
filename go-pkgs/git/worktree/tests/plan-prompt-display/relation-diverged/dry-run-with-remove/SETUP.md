# Scenario

**Feature**: diverged dry-run command listing

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DryRun = true
	req.CapturePrompt = false
	return nil
}
```