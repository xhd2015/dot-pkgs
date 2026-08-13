# Scenario

**Feature**: iTerm is resolvable but `OpenITerm` fails — no Terminal fallback

```
ResolveITerm -> fake iTerm.app
OpenITerm(dir) -> injected error
OpenConfig -> that error
OpenTerminal not called
```

## Steps

1. Set `OpenITermErr` to a distinctive injected message.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.OpenITermErr = "injected iterm2.Open failure"
	return nil
}
```
