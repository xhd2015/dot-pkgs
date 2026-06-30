# Scenario

**Feature**: `-r` reuse path targets current session without path scan or new tab

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-current")
	req.Mode = "reuse"
	return nil
}
```