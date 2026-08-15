# Scenario

**Feature**: reuse match branch selects matchingWindow before tab/session focus

```
BuildReuseCurrentSessionScript -> match: select matchingWindow -> select tab/session
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-match-window")
	req.Mode = "reuse"
	return nil
}
```