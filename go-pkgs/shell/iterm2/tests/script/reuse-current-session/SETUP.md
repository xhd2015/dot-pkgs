# Scenario

**Feature**: `-r` reuse path scans sessions; focus on match, new window + cd on miss

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-current")
	req.Mode = "reuse"
	return nil
}
```