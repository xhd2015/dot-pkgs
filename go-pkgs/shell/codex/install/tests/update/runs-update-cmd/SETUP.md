# Scenario

**Feature**: Update invokes RunShell once with UpdateCmd constant

```
Update(...) -> RunShell("codex update") once
```

## Steps

1. No extra flags; assert against `install.UpdateCmd`.

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
