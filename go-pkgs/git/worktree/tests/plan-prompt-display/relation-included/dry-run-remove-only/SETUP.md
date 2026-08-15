# Scenario

**Feature**: included branch dry-run — remove commands only

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DryRun = true
	return nil
}
```