# Scenario

**Feature**: iTerm resolve hit and `OpenITerm` succeeds

```
ResolveITerm -> /…/Applications/iTerm.app
OpenITerm(ValidDir) -> nil
OpenConfig -> {Via=iterm2, AppPath=that path}
OpenTerminal not called
```

## Steps

1. Leave `OpenITermErr` empty so the injected opener succeeds.
2. Do not set `TerminalApp` (unused on the iTerm path).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.OpenITermErr = ""
	return nil
}
```
