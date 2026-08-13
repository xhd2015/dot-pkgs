# Scenario

**Feature**: `OpenConfig(dir, cfg)` — validate then open via iTerm2 or Terminal

```
OpenConfig(dir, Config{ResolveITerm, OpenITerm, OpenTerminal, TerminalApp})
  -> reject | Via=iterm2 | Via=terminal
```

## Steps

1. Set `Operation=open-config`.
2. Default `Dir` to the existing `ValidDir` (reject leaves override).
3. Child groups set resolve / TerminalApp fixtures; leaves assert Result + spies.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "open-config"
	req.Dir = req.ValidDir
	return nil
}
```
