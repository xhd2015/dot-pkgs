# Scenario

**Feature**: Terminal opener error is returned; iTerm is not consulted

```
ResolveITerm -> ""
OpenTerminal(dir) -> injected error
OpenConfig -> that error
OpenITerm not called
```

## Steps

1. Set `OpenTerminalErr` to a distinctive injected message.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.OpenTerminalErr = "injected Terminal open failure"
	return nil
}
```
