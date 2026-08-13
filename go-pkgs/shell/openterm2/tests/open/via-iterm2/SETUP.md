# Scenario

**Feature**: resolvable iTerm2 → `OpenITerm` only (`Via=iterm2`)

```
ResolveITerm() -> fake iTerm.app
OpenConfig -> OpenITerm(dir) only
OpenTerminal never called
```

## Steps

1. Inject `ResolveITerm` to return a fake `.app` path under `WorkDir`.
2. Keep `Dir` as the existing `ValidDir`.
3. Leaves set opener success or injected `OpenITerm` error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermAppPath = fakeITermApp(req)
	req.Dir = req.ValidDir
	return nil
}
```
