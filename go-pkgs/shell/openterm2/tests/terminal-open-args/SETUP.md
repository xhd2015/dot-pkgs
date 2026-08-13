# Scenario

**Feature**: `TerminalOpenArgs(app, dir)` is the Terminal fallback argv (no exec)

```
TerminalOpenArgs(app, dir) -> []string{"open", "-a", app, absDir}
```

## Steps

1. Set `Operation=terminal-open-args`.
2. Default `Dir` to the existing absolute `ValidDir`.
3. Default `ArgsAppPath` to `/Applications/Utilities/Terminal.app`.
4. Leaves override app and/or dir; Assert compares the four-element argv.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "terminal-open-args"
	req.Dir = req.ValidDir
	req.ArgsAppPath = defaultTerminalApp
	return nil
}
```
