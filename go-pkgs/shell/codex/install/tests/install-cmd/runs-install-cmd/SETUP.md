# Scenario

**Feature**: Install invokes RunShell once with InstallCmd constant

```
Install(...) -> RunShell(`curl -fsSL https://chatgpt.com/codex/install.sh | sh`) once
```

## Steps

1. No extra flags; assert against `install.InstallCmd`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Happy path: no fail flags; Run records ShellCalls via injected RunShell.
	req.VersionCmdFail = false
	return nil
}
```
