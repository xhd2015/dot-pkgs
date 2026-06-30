# Scenario

**Feature**: `-r` reuse path scans sessions; focus on match, new window + cd on miss

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-current")
	req.Mode = "reuse"
	return nil
}
```