# Scenario

**Feature**: unknown detect — empty bash dump falls through to zsh

```
DetectShell -> ""
RunLogin("bash") <- empty
RunLogin("zsh")  <- FOO=zsh\0
ResolveLoginEnvs -> ("zsh", envs, nil)
```

## Steps

1. BashStdout empty (successful empty dump).
2. ZshStdout has FOO=zsh.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashStdout = ""
	req.ZshStdout = nulEnvDump("FOO=zsh")
	return nil
}
```
