# Scenario

**Feature**: zsh detect + multi-key dump → shell=zsh

```
DetectShell -> zsh
RunLogin("zsh") <- FOO=1\0GOPATH=/tmp/gp\0
ResolveLoginEnvs -> ("zsh", envs, nil)
```

## Steps

1. Inject ZshStdout with FOO=1 and GOPATH=/tmp/gp.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ZshStdout = nulEnvDump("FOO=1", "GOPATH=/tmp/gp")
	return nil
}
```
