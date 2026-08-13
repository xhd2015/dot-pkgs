# Scenario

**Feature**: Terminal fallback uses the default Terminal.app path

```
ResolveITerm -> ""
TerminalApp  -> "" (product default)
OpenConfig   -> {Via=terminal, AppPath=/Applications/Utilities/Terminal.app}
OpenITerm not called
```

## Steps

1. Leave `TerminalApp` empty so the product default applies.
2. Leave `OpenTerminalErr` empty so the injected opener succeeds.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.TerminalApp = ""
	req.OpenTerminalErr = ""
	return nil
}
```
