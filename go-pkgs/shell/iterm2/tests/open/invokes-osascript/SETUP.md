# Scenario

**Feature**: OpenConfig invokes injectable osascript with script

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = t.TempDir()
	return nil
}
```