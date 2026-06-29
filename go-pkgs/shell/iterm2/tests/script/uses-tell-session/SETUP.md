# Scenario

**Feature**: session path read via tell aSession

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-tell-session")
	return nil
}
```