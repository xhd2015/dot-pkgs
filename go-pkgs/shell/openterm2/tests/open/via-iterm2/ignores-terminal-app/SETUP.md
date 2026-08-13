# Scenario

**Feature**: `TerminalApp` does not change routing when iTerm is resolvable

```
ResolveITerm -> fake iTerm.app
TerminalApp  -> custom .app (must be ignored)
OpenConfig   -> Via=iterm2, AppPath=iTerm path
OpenTerminal not called
```

## Steps

1. Set `TerminalApp` to a custom path under `WorkDir`.
2. Leave `OpenITermErr` empty (success).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.TerminalApp = filepath.Join(req.WorkDir, "CustomTerminal.app")
	req.OpenITermErr = ""
	return nil
}
```
