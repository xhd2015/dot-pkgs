# Scenario

**Feature**: Terminal fallback uses `Config.TerminalApp` as `Result.AppPath`

```
ResolveITerm -> ""
TerminalApp  -> $WorkDir/CustomTerminal.app
OpenConfig   -> {Via=terminal, AppPath=CustomTerminal.app}
OpenITerm not called
```

## Steps

1. Set `TerminalApp` to a custom `.app` path under `WorkDir` (need not exist).
2. Leave `OpenTerminalErr` empty so the injected opener succeeds.

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
	req.OpenTerminalErr = ""
	return nil
}
```
