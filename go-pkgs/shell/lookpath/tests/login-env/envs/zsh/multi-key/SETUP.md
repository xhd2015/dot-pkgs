# Scenario

**Feature**: zsh Envs parse multi-key NUL dump; FOO and GOPATH present

```
RunLogin("zsh", env -0, HOME=…)
  <- inject FOO=1\0GOPATH=/tmp/gp\0
ResolveZshLoginEnvs -> []string includes both KEY=value
```

## Steps

1. Set Home under WorkDir for spy.
2. Inject LoginStdout as NUL-delimited dump with FOO=1 and GOPATH=/tmp/gp.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Home = filepath.Join(req.WorkDir, "home")
	req.LoginStdout = nulEnvDump("FOO=1", "GOPATH=/tmp/gp")
	return nil
}
```
