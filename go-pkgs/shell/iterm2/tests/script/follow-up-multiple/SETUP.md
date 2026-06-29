# Scenario

**Feature**: multiple follow-ups preserved in order

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-follow-multi")
	req.FollowUps = []string{"grok", "codex"}
	return nil
}
```