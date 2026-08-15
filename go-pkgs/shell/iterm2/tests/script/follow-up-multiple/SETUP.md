# Scenario

**Feature**: multiple follow-ups preserved in order

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-follow-multi")
	req.FollowUps = []string{"grok", "codex"}
	return nil
}
```