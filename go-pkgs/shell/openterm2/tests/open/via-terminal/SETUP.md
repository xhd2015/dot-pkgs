# Scenario

**Feature**: iTerm not resolvable → `OpenTerminal` only (`Via=terminal`)

```
ResolveITerm() -> ""
OpenConfig -> OpenTerminal(dir) only
OpenITerm never called
```

## Steps

1. Inject `ResolveITerm` to return `""`.
2. Keep `Dir` as the existing `ValidDir`.
3. Leaves set default vs override `TerminalApp`, or injected Terminal error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermAppPath = ""
	req.Dir = req.ValidDir
	return nil
}
```
