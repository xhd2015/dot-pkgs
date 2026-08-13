# Scenario

**Feature**: relative `dir` is converted to an absolute path in argv

```
TerminalOpenArgs(default Terminal.app, "rel-project")
  -> open -a /Applications/Utilities/Terminal.app <filepath.Abs("rel-project")>
```

## Steps

1. Set `Dir` to a relative path (`rel-project`). Do not `Chdir`.
2. Keep default `ArgsAppPath`.
3. Assert uses `filepath.Abs` in the same process so cwd need not be known.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Dir = "rel-project"
	req.ArgsAppPath = defaultTerminalApp
	return nil
}
```
