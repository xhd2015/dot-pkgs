# Scenario

**Feature**: non-bash/zsh detect (e.g. fish) uses the unknown bash→zsh cascade

```
DetectShell -> "fish"
RunLogin("bash") <- empty
RunLogin("zsh")  <- FOO=from-other\0
ResolveLoginEnvs -> ("zsh", envs, nil)
```

## Steps

1. Set DetectShellResult to `fish` (not bash/zsh).
2. Bash empty; zsh dump nonempty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DetectShellResult = "fish"
	req.BashStdout = ""
	req.ZshStdout = nulEnvDump("FOO=from-other")
	return nil
}
```
