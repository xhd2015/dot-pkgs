# Scenario

**Feature**: single follow-up command after cd

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-follow-one")
	req.FollowUps = []string{"grok"}
	return nil
}
```