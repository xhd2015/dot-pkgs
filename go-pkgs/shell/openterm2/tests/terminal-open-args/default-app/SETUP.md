# Scenario

**Feature**: default Terminal.app argv for an absolute directory

```
TerminalOpenArgs(/Applications/Utilities/Terminal.app, ValidDir)
  -> open -a /Applications/Utilities/Terminal.app <abs ValidDir>
```

## Steps

1. Keep grouping defaults: default app + absolute `ValidDir`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ArgsAppPath = defaultTerminalApp
	req.Dir = req.ValidDir
	return nil
}
```
