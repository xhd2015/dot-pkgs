# Scenario

**Feature**: smart-open match branch scopes cd to matchingWindow's new tab

```
BuildScript -> match: create tab in matchingWindow -> cd in that tab (not frontmost)
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-smart-match-cd")
	return nil
}
```