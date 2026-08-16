# Scenario

**Feature**: unknown detect — bash RunLogin error falls through to zsh

```
DetectShell -> ""
RunLogin("bash") -> error
RunLogin("zsh")  <- FOO=zsh\0
ResolveLoginEnvs -> ("zsh", envs, nil)
```

## Steps

1. BashFail=true.
2. ZshStdout has FOO=zsh.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashFail = true
	req.ZshStdout = nulEnvDump("FOO=zsh")
	return nil
}
```
