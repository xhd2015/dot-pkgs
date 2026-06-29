# Scenario

**Feature**: login shell must not be replaced with one-shot exec

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-no-exec")
	return nil
}
```