# Scenario

**Feature**: unknown detect + nonempty bash dump short-circuits (zsh not called)

```
DetectShell -> ""
RunLogin("bash") <- FOO=1\0GOPATH=/tmp/gp\0
ResolveLoginEnvs -> ("bash", envs, nil); zsh not invoked
```

## Steps

1. DetectShellResult remains empty (parent).
2. Inject BashStdout multi-key dump; leave zsh empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashStdout = nulEnvDump("FOO=1", "GOPATH=/tmp/gp")
	return nil
}
```
