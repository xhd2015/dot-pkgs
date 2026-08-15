# Scenario

**Feature**: single follow-up command after cd

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-follow-one")
	req.FollowUps = []string{"grok"}
	return nil
}
```