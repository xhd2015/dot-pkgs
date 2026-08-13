# Scenario

**Feature**: argv uses the override app path, not the default Terminal.app

```
TerminalOpenArgs($WorkDir/CustomTerminal.app, ValidDir)
  -> open -a $WorkDir/CustomTerminal.app <abs ValidDir>
```

## Steps

1. Set `ArgsAppPath` to a custom `.app` under `WorkDir`.
2. Keep `Dir` as the absolute `ValidDir`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ArgsAppPath = filepath.Join(req.WorkDir, "CustomTerminal.app")
	req.Dir = req.ValidDir
	return nil
}
```
